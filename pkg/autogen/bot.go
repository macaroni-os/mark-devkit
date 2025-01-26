/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/macaroni-os/mark-devkit/pkg/autogen/generators"
	tmpleng "github.com/macaroni-os/mark-devkit/pkg/autogen/tmpl-engines"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type AutogenBot struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	WorkDir     string
	DownloadDir string

	GithubClient *github.Client
	RestGuard    *guard.RestGuard

	// Merge section
	MergeBot  *kit.MergeBot
	MergeOpts *kit.MergeBotOpts

	Fetcher kit.Fetcher

	ElabAtoms []*specs.RepoScanAtom
	mutex     sync.Mutex
}

type AutogenBotOpts struct {
	PullSources         bool
	Push                bool
	PullRequest         bool
	Verbose             bool
	CleanWorkingDir     bool
	GenReposcan         bool
	SyncFiles           bool
	MergeAutogen        bool
	MergeForced         bool
	ShowGeneratedValues bool
	Atoms               []string

	GitDeepFetch int
	Concurrency  int

	SignatureName  string
	SignatureEmail string

	// Pull Request data
	GithubUser string
}

func NewAutogenBotOpts() *AutogenBotOpts {
	return &AutogenBotOpts{
		Push:                true,
		PullSources:         true,
		PullRequest:         false,
		SyncFiles:           true,
		Concurrency:         10,
		CleanWorkingDir:     true,
		GithubUser:          "macaroni-os",
		MergeAutogen:        true,
		MergeForced:         true,
		ShowGeneratedValues: false,
		Atoms:               []string{},
	}
}

func (o *AutogenBotOpts) HasAtoms() bool {
	return len(o.Atoms) > 0
}

func (o *AutogenBotOpts) AtomInFilter(atom string) bool {
	for _, name := range o.Atoms {
		if name == atom {
			return true
		}
	}
	return false
}

func NewAutogenBot(c *specs.MarkDevkitConfig) *AutogenBot {
	rcfg := guard_specs.NewConfig()
	rg, _ := guard.NewRestGuard(rcfg)
	// Overide the default check redirect
	rg.Client.CheckRedirect = kit.CheckRedirect

	return &AutogenBot{
		Config:       c,
		Logger:       log.GetDefaultLogger(),
		WorkDir:      "./workdir",
		GithubClient: nil,
		RestGuard:    rg,
		ElabAtoms:    []*specs.RepoScanAtom{},
	}
}

func (a *AutogenBot) GetTargetDir() string {
	return filepath.Join(a.WorkDir, "dest")
}

func (a *AutogenBot) GetDownloadDir() string {
	if a.DownloadDir != "" {
		return a.DownloadDir
	}
	return filepath.Join(a.WorkDir, "downloads")
}

func (a *AutogenBot) GetSourcesDir() string {
	return filepath.Join(a.WorkDir, "sources")
}

func (a *AutogenBot) SetWorkDir(d string)     { a.WorkDir = d }
func (a *AutogenBot) SetDownloadDir(d string) { a.DownloadDir = d }

func (a *AutogenBot) SetupGithubClient(ctx context.Context) error {
	if a.GithubClient == nil {
		pushOpts := kit.NewPushOptions()

		remote, available := a.Config.GetAuthentication().GetRemote("github.com")
		if available {
			pushOpts.Token = remote.Token
		}

		auth, err := kit.GetGithubAuth(pushOpts)
		if err != nil {
			return err
		}

		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: auth.Password,
		})
		tc := oauth2.NewClient(ctx, ts)
		a.GithubClient = github.NewClient(tc)
	}

	return nil
}

func (a *AutogenBot) SetupFetcher(backend string, opts map[string]string) error {
	var err error
	a.Fetcher, err = kit.NewFetcher(a.Config, backend, opts)
	if err != nil {
		return err
	}

	a.Fetcher.SetWorkDir(a.WorkDir)
	a.Fetcher.SetDownloadDir(a.DownloadDir)

	return err
}

func (a *AutogenBot) AddReposcanAtom(atom *specs.RepoScanAtom) {
	defer a.mutex.Unlock()
	a.mutex.Lock()
	a.ElabAtoms = append(a.ElabAtoms, atom)
}

func (a *AutogenBot) setupMergeBot(mkit *specs.MergeKit, opts *AutogenBotOpts) error {
	a.MergeOpts = kit.NewMergeBotOpts()
	a.MergeOpts.Concurrency = opts.Concurrency
	a.MergeOpts.GenReposcan = opts.GenReposcan
	a.MergeOpts.PullRequest = opts.PullRequest
	a.MergeOpts.CleanWorkingDir = false
	a.MergeOpts.Push = opts.Push
	a.MergeOpts.GithubUser = opts.GithubUser
	a.MergeOpts.SignatureEmail = opts.SignatureEmail
	a.MergeOpts.SignatureName = opts.SignatureName
	a.MergeOpts.GitDeepFetch = opts.GitDeepFetch
	a.MergeOpts.Verbose = opts.Verbose

	a.MergeBot = kit.NewMergeBot(a.Config)
	a.MergeBot.SetWorkDir(a.WorkDir)

	_, err := mkit.GetTargetKit()
	if err != nil {
		return err
	}

	if opts.PullSources {
		err = a.MergeBot.CloneSourcesKits(mkit, a.MergeOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AutogenBot) setupTargetKit(mkit *specs.MergeKit, opts *AutogenBotOpts) error {
	targetKit, _ := mkit.GetTargetKit()

	// Clone target kit
	err := a.MergeBot.CloneTargetKit(mkit, a.MergeOpts)
	if err != nil {
		return err
	}

	if opts.GenReposcan {
		a.Logger.InfoC(a.Logger.Aurora.Bold(
			fmt.Sprintf(":brain:[%s] Generating reposcan files...",
				targetKit.Name)))

		err = a.MergeBot.GenerateReposcanFiles(mkit, a.MergeOpts, false)
		if err != nil {
			return err
		}
	}

	if !a.MergeBot.TargetKitIsANewBranch() {
		err = a.MergeBot.SetupTargetResolver(mkit, a.MergeOpts, targetKit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AutogenBot) UpdateTargetKit(mkit *specs.MergeKit, candidates []*specs.RepoScanAtom,
	opts *AutogenBotOpts) error {

	// Copy eclasses
	err := a.MergeBot.MergeEclasses(mkit, a.MergeOpts)
	if err != nil {
		return err
	}

	// Prepare metadata directory of the kit
	err = a.MergeBot.PrepareMetadataDir(mkit, a.MergeOpts)
	if err != nil {
		return err
	}

	// Prepare profiles directory of the kit
	err = a.MergeBot.PrepareProfilesDir(mkit, candidates, a.MergeOpts)
	if err != nil {
		return err
	}

	// Copy fixups
	err = a.MergeBot.MergeFixups(mkit, a.MergeOpts)
	if err != nil {
		return err
	}

	return nil
}

func (a *AutogenBot) Run(specfile, kitFile string, opts *AutogenBotOpts) error {
	// Load Autogen specs
	aspec := specs.NewAutogenSpec()
	mkit := specs.NewMergeKit()
	ctx := context.Background()

	if opts.CleanWorkingDir {
		defer os.RemoveAll(a.WorkDir)
	}

	err := aspec.LoadFile(specfile)
	if err != nil {
		return err
	}
	aspec.Prepare()

	err = mkit.LoadFile(kitFile)
	if err != nil {
		return err
	}

	// Setup merge bot and target kit
	err = a.setupMergeBot(mkit, opts)
	if err != nil {
		return err
	}

	targetKit, _ := mkit.GetTargetKit()

	a.Logger.InfoC(a.Logger.Aurora.Bold(
		fmt.Sprintf(":castle:Work directory:\t%s\n:rocket:Target Kit:\t\t%s",
			a.WorkDir, targetKit.Name)))

	// Setup target kit
	if opts.MergeAutogen {
		err = a.setupTargetKit(mkit, opts)
		if err != nil {
			return err
		}
	}

	// NOTE: This must be done only if there are
	//       generator using github or for pull requests
	if aspec.HasGithubGenerators() || opts.PullRequest {
		err = a.SetupGithubClient(ctx)
		if err != nil {
			return err
		}

		a.MergeBot.GithubClient = a.GithubClient
	}

	// Process definitions
	err = a.ProcessDefinitions(mkit, aspec, opts)
	if err != nil {
		return err
	}

	// Prepare resolver
	err = a.prepareResolver(mkit, aspec, opts)
	if err != nil {
		return err
	}

	if !opts.MergeAutogen {
		// Stop processing
		return nil
	}

	// Bump packages
	err = a.MergeBot.ElaborateMerge(mkit, a.MergeOpts, targetKit)
	if err != nil {
		return err
	}

	// Sync download files on selected backend.
	return a.syncTarballs(opts)
}

func (a *AutogenBot) prepareResolver(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, opts *AutogenBotOpts) error {

	// Create AutogenSpec for all elaborated atoms.
	s := specs.RepoScanSpec{
		CacheDataVersion: specs.CacheDataVersion,
		Atoms:            make(map[string]specs.RepoScanAtom, 0),
		MetadataErrors:   make(map[string]specs.RepoScanAtom, 0),
		File:             "autogen",
	}

	for idx := range a.ElabAtoms {
		s.Atoms[a.ElabAtoms[idx].Atom] = *a.ElabAtoms[idx]
	}

	a.MergeBot.Resolver.Sources = append(a.MergeBot.Resolver.Sources, s)

	return a.MergeBot.Resolver.BuildMap()
}

func (a *AutogenBot) ProcessDefinitions(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, opts *AutogenBotOpts) error {

	// Prepare download dir
	err := helpers.EnsureDirWithoutIds(a.GetDownloadDir(), 0755)
	if err != nil {
		return err
	}

	// Prepare staging dir
	err = helpers.EnsureDirWithoutIds(a.GetSourcesDir(), 0755)
	if err != nil {
		return err
	}

	for name, def := range aspec.Definitions {
		a.Logger.Info(fmt.Sprintf(
			":factory:[%s] Processing definition ...", name))

		if len(def.Packages) == 0 {
			continue
		}

		if def.Generator == "" {
			a.Logger.Info(fmt.Sprintf(
				":warning:[%s] Generator not defined. Ignoring definition.",
				name))
			continue
		}

		err = a.ProcessDefinition(mkit, aspec, def, opts, name)
		if err != nil {
			return err
		}

	}

	return nil
}

func (a *AutogenBot) ProcessDefinition(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, def *specs.AutogenDefinition,
	opts *AutogenBotOpts, nameDef string) error {

	// Prepare generator
	generator, err := a.GetGenerator(def.Generator)
	if err != nil {
		return err
	}

	// Prepare template engine
	templateEngine, err := a.GetTemplateEngine(def.TemplateEngine)
	if err != nil {
		return err
	}

	atomFiltered := opts.HasAtoms()

	// Process packages
	for _, pkg := range def.Packages {
		for _, atom := range pkg {

			if atomFiltered && !opts.AtomInFilter(atom.Name) {
				a.Logger.Debug(fmt.Sprintf(
					":factory:[%s] Atom %s filtered.", nameDef, atom.Name))
				continue
			}

			a.Logger.Info(fmt.Sprintf(
				":factory:[%s] Processing atom %s...", nameDef, atom.Name))

			err = a.ProcessPackage(mkit, aspec, atom, def.Defaults, generator, templateEngine, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *AutogenBot) GetGenerator(generatorType string) (generators.Generator, error) {
	ans, err := generators.NewGenerator(generatorType)
	if err != nil {
		return ans, err
	}

	if ans.GetType() == specs.GeneratorBuiltinGitub {
		gg := ans.(*generators.GithubGenerator)
		gg.SetClient(a.GithubClient)
	}

	return ans, nil
}

func (a *AutogenBot) GetTemplateEngine(t *specs.AutogenTemplateEngine) (tmpleng.TemplateEngine, error) {
	engine := "helm"
	opts := []string{}
	if t != nil {
		engine = t.Engine
		opts = t.Opts
	}
	ans, err := tmpleng.NewTemplateEngine(engine, opts)
	if err != nil {
		return nil, err
	}

	ans.SetLogger(a.Logger)
	return ans, nil
}

func (a *AutogenBot) ProcessPackage(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom,
	generator generators.Generator, tmplEngine tmpleng.TemplateEngine,
	opts *AutogenBotOpts) error {

	if def == nil {
		// Create an empty default object if not present.
		def = &specs.AutogenAtom{
			Github: &specs.AutogenGithubProps{},
			Dir:    &specs.AutogenDirlistingProps{},
			Python: &specs.AutogenPythonOpts{},
			Vars:   make(map[string]interface{}, 0),
		}
	} else {
		// Ensure that object are not nil to avoid segfault.
		if def.Github == nil {
			def.Github = &specs.AutogenGithubProps{}
		}
		if def.Dir == nil {
			def.Dir = &specs.AutogenDirlistingProps{}
		}
		if def.Python == nil {
			def.Python = &specs.AutogenPythonOpts{}
		}
	}

	def = def.Clone()
	atom = def.Merge(atom)

	// Retrieve package metadata and last versions
	valuesRef, err := generator.Process(atom)
	if err != nil {
		return err
	}
	values := *valuesRef

	mergeF := func(vars *map[string]interface{}) error {
		for k, v := range *vars {
			if k == "versions" {

				ilist, ok := v.([]interface{})
				if !ok {
					return fmt.Errorf(
						"Invalid type on versions var for package %s",
						atom.Name)
				}

				// Special case.
				// I need to convert []interface{} to []string
				vlist := []string{}
				for _, vv := range ilist {
					str, ok := vv.(string)
					if !ok {
						return fmt.Errorf(
							"Invalid value %v on versions var for package %s",
							vv, atom.Name)
					}
					vlist = append(vlist, str)
				}
				values[k] = vlist
			} else {
				values[k] = v
			}
		}

		return nil
	}

	if len(def.Vars) > 0 {
		err = mergeF(&def.Vars)
		if err != nil {
			return err
		}
	}

	if len(atom.Vars) > 0 {
		err = mergeF(&atom.Vars)
		if err != nil {
			return err
		}
	}

	versionsI, present := values["versions"]
	if !present {
		return fmt.Errorf("No versions found for package %s", atom.Name)
	}
	versions, _ := versionsI.([]string)

	// Sanitize versions or use them directly
	sanitizedVersions := []string{}
	var vMap *map[string]string
	if atom.HasTransforms() {

		if opts.ShowGeneratedValues {
			data, err := yaml.Marshal(values)
			if err != nil {
				return err
			}
			a.Logger.InfoC(fmt.Sprintf(
				":eyes:[%s] Values Before Transforms:\n%s", atom.Name,
				string(data)))
		}

		vMap, err = a.transformsVersions(atom, versions)
		if err != nil {
			return err
		}
		for _, sv := range *vMap {
			sanitizedVersions = append(sanitizedVersions, sv)
		}

		sanitizedVersions, err = a.sortVersions(atom, sanitizedVersions)
		if err != nil {
			return err
		}

	} else {
		sanitizedVersions, err = a.sortVersions(atom, versions)
		if err != nil {
			return err
		}
	}

	if len(sanitizedVersions) == 0 {
		return fmt.Errorf("[%s] No versions found", atom.Name)
	}

	// Select version
	selectedVersion := ""
	if atom.HasSelector() {
		selectedVersion, err = a.selectVersion(atom, def, sanitizedVersions)
		if err != nil {
			return err
		}
	} else {
		selectedVersion = sanitizedVersions[0]
	}

	a.Logger.Info(fmt.Sprintf(
		":pizza:[%s] For package %s/%s selected version %s",
		atom.Name, atom.GetCategory(def), atom.Name, selectedVersion))

	values["version"] = selectedVersion
	values["category"] = atom.GetCategory(def)
	values["original_version"] = selectedVersion
	values["pn"] = atom.Name
	if atom.HasTransforms() {
		// Retrieve original version
		for v, sv := range *vMap {
			if sv == selectedVersion {
				values["original_version"] = v
				break
			}
		}
	}

	if opts.ShowGeneratedValues {
		data, err := yaml.Marshal(values)
		if err != nil {
			return err
		}
		a.Logger.InfoC(fmt.Sprintf(
			":eyes:[%s] Values:\n%s", atom.Name,
			string(data)))
	}

	// Prepare metadata of the selected version
	err = generator.SetVersion(atom, selectedVersion, &values)
	if err != nil {
		return err
	}

	if opts.ShowGeneratedValues {
		data, err := yaml.Marshal(values)
		if err != nil {
			return err
		}
		a.Logger.InfoC(fmt.Sprintf(
			":eyes:[%s] Values for templates:\n%s", atom.Name,
			string(data)))
	}

	toAdd, err := a.isVersion2Add(
		atom, def, selectedVersion, opts)
	if err != nil {
		return err
	}

	if !toAdd {
		a.Logger.InfoC(fmt.Sprintf(
			":smiling_face_with_sunglasses:[%s] Package already present.",
			atom.Name))
		return nil
	}

	// Download artefacts and prepare stagings dir.
	reposcanAtom, err := a.GeneratePackageOnStaging(
		mkit, aspec, atom, def, &values, tmplEngine)
	if err != nil {
		return err
	}

	// Add reposcan to elab list
	a.AddReposcanAtom(reposcanAtom)

	return nil
}
