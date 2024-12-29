/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sync"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/go-git/go-git/v5"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type FetcherCommon struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	Resolver *RepoScanResolver

	WorkDir string

	Stats     *AtomsStats
	RestGuard *guard.RestGuard

	AtomInError []*AtomError
	mutex       sync.Mutex
}

type AtomError struct {
	Atom  *specs.RepoScanAtom `json:"atom,omitempty" yaml:"atom,omitempty"`
	Error string              `json:"error,omitempty" yaml:"error,omitempty"`
}

type AtomsStats struct {
	TotAtoms  int `json:"tot_atoms,omitempty" yaml:"tot_atoms,omitempty"`
	TotErrors int `json:"tot_errors,omitempty" yaml:"tot_errors,omitempty"`

	mutex sync.Mutex `json:"-" yaml:"-"`
}

func NewAtomsStats() *AtomsStats {
	return &AtomsStats{
		TotAtoms:  0,
		TotErrors: 0,
	}
}

func (a *AtomsStats) IncrementErrors() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.TotErrors++
}

func (a *AtomsStats) IncrementAtoms() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.TotAtoms++
}

func CheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	// Workaround to sourceforge download. We need to drop Referer
	req.Header.Del("Referer")

	return nil
}

func NewFetcherCommon(c *specs.MarkDevkitConfig) *FetcherCommon {
	resolver := NewRepoScanResolver(c)
	rcfg := guard_specs.NewConfig()
	//rcfg.DisableCompression = true
	rg, _ := guard.NewRestGuard(rcfg)
	// Overide the default check redirect
	rg.Client.CheckRedirect = CheckRedirect
	return &FetcherCommon{
		Config:      c,
		Logger:      resolver.Logger,
		Resolver:    resolver,
		WorkDir:     "./workdir",
		RestGuard:   rg,
		Stats:       NewAtomsStats(),
		AtomInError: []*AtomError{},
	}
}

func (f *FetcherCommon) GetTargetDir() string {
	return filepath.Join(f.WorkDir, "kits")
}

func (f *FetcherCommon) GetReposcanDir() string {
	return filepath.Join(f.WorkDir, "kit-cache")
}

func (f *FetcherCommon) GetDownloadDir() string {
	return filepath.Join(f.WorkDir, "download")
}

func (f *FetcherCommon) GetStats() *AtomsStats { return f.Stats }
func (f *FetcherCommon) SetWorkDir(d string)   { f.WorkDir = d }
func (f *FetcherCommon) GetWorkDir() string    { return f.WorkDir }

func (f *FetcherCommon) AddAtomInError(a *specs.RepoScanAtom, err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.AtomInError = append(f.AtomInError,
		&AtomError{
			Atom:  a,
			Error: err.Error(),
		})
}

func (f *FetcherCommon) GetAtomsInError() *[]*AtomError {
	return &f.AtomInError
}

func (f *FetcherCommon) PrepareSourcesKits(mkit *specs.MergeKit, opts *FetchOpts) error {
	gitOpts := &CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			SingleBranch: true,
			RemoteName:   "origin",
			Depth:        5,
		},
		Verbose: opts.Verbose,
		// Always generate summary report
		Summary: false,
		Results: []*specs.ReposcanKit{},
	}

	analysis := &specs.ReposcanAnalysis{
		Kits: mkit.Sources,
	}

	// Clone sources kit
	err := CloneKits(analysis, f.GetTargetDir(), gitOpts)
	if err != nil {
		return err
	}

	if opts.GenReposcan {

		err = f.GenerateReposcanFiles(mkit, opts)
		if err != nil {
			return err
		}
	}

	// Check if the reposcan files are present and
	// prepare resolver
	for _, source := range mkit.Sources {
		targetFile := filepath.Join(f.GetReposcanDir(), source.Name+"-"+source.Branch)
		if !utils.Exists(targetFile) {
			return fmt.Errorf("Cache file %s-%s not found. Generate it!",
				source.Name, source.Branch)
		}
		f.Resolver.JsonSources = append(f.Resolver.JsonSources, targetFile)
	}

	// Load reposcan files
	err = f.Resolver.LoadJsonFiles(opts.Verbose)
	if err != nil {
		return err
	}

	// Build resolver map
	err = f.Resolver.BuildMap()
	if err != nil {
		return err
	}

	return nil
}

func (f *FetcherCommon) GenerateReposcanFiles(mkit *specs.MergeKit, opts *FetchOpts) error {
	err := helpers.EnsureDirWithoutIds(f.GetReposcanDir(), 0755)
	if err != nil {
		return err
	}

	// Prepare eclass dir list
	eclassDirs := []string{}
	for _, source := range mkit.Sources {
		eclassDir, err := filepath.Abs(filepath.Join(f.GetTargetDir(), source.Name, "eclass"))
		if err != nil {
			return err
		}
		if utils.Exists(eclassDir) {
			kitDir, _ := filepath.Abs(filepath.Join(f.GetTargetDir(), source.Name))
			eclassDirs = append(eclassDirs, kitDir)
		}
	}

	for _, source := range mkit.Sources {
		sourceDir := filepath.Join(f.GetTargetDir(), source.Name)
		targetFile := filepath.Join(f.GetReposcanDir(), source.Name+"-"+source.Branch)

		err := f.GenerateKitCacheFile(sourceDir, source.Name, source.Branch,
			targetFile, eclassDirs, opts.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FetcherCommon) GenerateKitCacheFile(sourceDir, kitName, kitBranch, targetFile string,
	eclassDirs []string, concurrency int) error {
	f.Logger.Debug(fmt.Sprintf("Generating kit-cache file for kit %s...",
		kitName))

	return RunReposcanGenerate(sourceDir, kitName, kitBranch, targetFile,
		eclassDirs, concurrency)
}

func (f *FetcherCommon) DownloadAtomsFiles(mkit *specs.MergeKit, atom *specs.RepoScanAtom) error {
	var err error
	var uri *url.URL

	fileNamesMap := make(map[string]bool, 0)

	for _, file := range atom.Files {

		if _, present := fileNamesMap[file.Name]; present {
			continue
		}

		uri, err = url.Parse(file.SrcUri[0])
		if err != nil {
			return err
		}

		file512, _ := file.Hashes["sha512"]
		fileBlake2b, _ := file.Hashes["blake2b"]
		atomUrl := file.SrcUri[0]

		if uri.Scheme == "mirror" {
			uris := mkit.Target.GetThirdpartyMirrorsUris(uri.Host)

			if len(uris) == 0 {
				return fmt.Errorf("No mirrors urls found for alias %s",
					uri.Host)
			}

			for _, mirrorUri := range uris {

				if mirrorUri[len(mirrorUri)-1:len(mirrorUri)] == "/" {
					atomUrl = mirrorUri[:len(mirrorUri)-1] + uri.Path
				} else {
					atomUrl = mirrorUri + uri.Path
				}

				err = f.downloadArtefact(atomUrl, file.Name, file512, fileBlake2b)
				if err == nil {
					break
				}

				if f.Config.GetGeneral().Debug {

					f.Logger.Info(fmt.Sprintf(":cross_mark:[%s] (%s) %s - %s: %s",
						atom.Atom, uri.Host, atomUrl, path.Base(atomUrl), err.Error()))
				} else {
					f.Logger.Info(fmt.Sprintf(":cross_mark:[%s] (%s) %s - %s",
						atom.Atom, uri.Host, atomUrl, path.Base(atomUrl)))
				}
			}

		} else {
			err = f.downloadArtefact(atomUrl, file.Name, file512, fileBlake2b)
		}

		if err != nil {
			break
		}

		f.Logger.Info(fmt.Sprintf(":check_mark: [%s] %s - %s",
			atom.Atom, atomUrl, path.Base(atomUrl)))

		fileNamesMap[file.Name] = true
	}

	return err
}

func (f *FetcherCommon) downloadArtefact(atomUrl, atomName, fileSha512, fileBlake2b string) error {

	uri, err := url.Parse(atomUrl)
	if err != nil {
		return err
	}

	ssl := false

	switch uri.Scheme {
	case "https":
		ssl = true
	default:
		ssl = false
	}

	if uri.Scheme == "ftp" {
		return fmt.Errorf("Not yet implemented!")
	} else {

		node := guard_specs.NewRestNode(uri.Host,
			uri.Host+path.Dir(uri.Path), ssl)

		resource := path.Base(uri.Path)

		service := guard_specs.NewRestService(uri.Host)
		service.Retries = 3
		service.AddNode(node)

		t := service.GetTicket()
		defer t.Rip()

		_, err := f.RestGuard.CreateRequest(t, "GET", "/"+resource)
		if err != nil {
			return err
		}

		downloadedFilePath := filepath.Join(f.GetDownloadDir(), atomName)

		artefact, err := f.RestGuard.DoDownload(t, downloadedFilePath)
		if err != nil {
			if t.Response != nil {
				return fmt.Errorf("%s - %s - %s", atomName, err.Error(), t.Response.Status)
			} else {
				return fmt.Errorf("%s - %s", atomName, err.Error())
			}
		}

		if artefact.Sha512 != fileSha512 {
			return fmt.Errorf("file %s with sha512 %s instead of %s",
				atomName, artefact.Sha512, fileSha512)
		}

		if fileBlake2b != "" && artefact.Blake2b != fileBlake2b {
			return fmt.Errorf("file %s with blake2b %s instead of %s",
				atomName, artefact.Sha512, fileBlake2b)
		}

	}

	return nil
}