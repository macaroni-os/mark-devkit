/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/go-git/go-git/v5"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type MergeBot struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	Resolver       *RepoScanResolver
	TargetResolver *RepoScanResolver

	IsANewBranch bool
	WorkDir      string

	hasCommit     bool
	files4Commit  map[string][]string
	manifestFiles map[string][]specs.RepoScanFile
}

type MergeBotOpts struct {
	GenReposcan     bool
	DryRun          bool
	PullSources     bool
	Push            bool
	PullRequest     bool
	Verbose         bool
	CleanWorkingDir bool

	GitDeepFetch int
	Concurrency  int

	SignatureName  string
	SignatureEmail string
}

func NewMergeBotOpts() *MergeBotOpts {
	return &MergeBotOpts{
		GenReposcan:     true,
		DryRun:          false,
		Push:            true,
		PullSources:     true,
		PullRequest:     false,
		GitDeepFetch:    10,
		Concurrency:     10,
		CleanWorkingDir: true,
	}
}

func NewMergeBot(c *specs.MarkDevkitConfig) *MergeBot {
	resolver := NewRepoScanResolver(c)
	targetResolver := NewRepoScanResolver(c)
	return &MergeBot{
		Config:         c,
		Logger:         resolver.Logger,
		Resolver:       resolver,
		TargetResolver: targetResolver,
		IsANewBranch:   false,
		WorkDir:        "./workdir",
		hasCommit:      false,
		files4Commit:   make(map[string][]string, 0),
		manifestFiles:  make(map[string][]specs.RepoScanFile, 0),
	}
}

func (m *MergeBot) GetSourcesDir() string {
	return filepath.Join(m.WorkDir, "sources")
}

func (m *MergeBot) GetTargetDir() string {
	return filepath.Join(m.WorkDir, "dest")
}

func (m *MergeBot) GetReposcanDir() string {
	return filepath.Join(m.WorkDir, "kit-cache")
}

func (m *MergeBot) GetResolver() *RepoScanResolver { return m.Resolver }
func (m *MergeBot) SetWorkDir(d string)            { m.WorkDir = d }

func (m *MergeBot) Run(specfile string, opts *MergeBotOpts) error {
	// Load MergeKit specs
	mkit := specs.NewMergeKit()

	defer os.RemoveAll(m.WorkDir)

	err := mkit.LoadFile(specfile)
	if err != nil {
		return err
	}

	targetKit, err := mkit.GetTargetKit()
	if err != nil {
		return err
	}

	m.Logger.InfoC(m.Logger.Aurora.Bold(
		fmt.Sprintf(":castle:Work directory:\t%s\n:rocket:Target Kit:\t\t%s",
			m.WorkDir, targetKit.Name)))

	if opts.PullSources {
		// Clone sources
		err = m.cloneSourcesKits(mkit, opts)
		if err != nil {
			return err
		}
	}

	// Clone target kit
	err = m.cloneTargetKit(mkit, opts)
	if err != nil {
		return err
	}

	// Generate kit-cache files
	if opts.GenReposcan {

		m.Logger.InfoC(m.Logger.Aurora.Bold(
			fmt.Sprintf(":brain:[%s] Generating reposcan files...",
				targetKit.Name)))
		err = m.GenerateReposcanFiles(mkit, opts)
		if err != nil {
			return err
		}

	}

	// Check if the reposcan files are present and
	// prepare resolver
	for _, source := range mkit.Sources {
		targetFile := filepath.Join(m.GetReposcanDir(), source.Name+"-"+source.Branch)
		if !utils.Exists(targetFile) {
			return fmt.Errorf("Cache file %s-%s not found. Generate it!",
				source.Name, source.Branch)
		}
		m.Resolver.JsonSources = append(m.Resolver.JsonSources, targetFile)
	}

	// Load reposcan files
	err = m.Resolver.LoadJsonFiles(opts.Verbose)
	if err != nil {
		return err
	}

	// Build resolver map
	err = m.Resolver.BuildMap()
	if err != nil {
		return err
	}

	// Prepare target resolver
	if !m.IsANewBranch {
		m.TargetResolver.JsonSources = []string{
			filepath.Join(m.GetReposcanDir(),
				"target-"+targetKit.Name+"-"+targetKit.Branch)}

		// Load reposcan files
		err = m.TargetResolver.LoadJsonFiles(opts.Verbose)
		if err != nil {
			return err
		}

		// Build resolver map
		err = m.TargetResolver.BuildMap()
		if err != nil {
			return err
		}
	}

	// Search Atoms
	candidates, err := m.SearchAtoms(mkit, opts)
	if err != nil {
		return err
	}

	if len(candidates) > 0 {
		m.Logger.Info(fmt.Sprintf(":dart:Found %d candidates:",
			len(candidates)))

		for _, candidate := range candidates {
			m.Logger.Info(fmt.Sprintf(":pizza:[%s] %s",
				candidate.Kit, candidate.Atom))
		}

		// Merge Atoms
		err = m.MergeAtoms(candidates, mkit, opts)
		if err != nil {
			return err
		}

		if m.Config.GetGeneral().Debug {
			for k, files := range m.files4Commit {
				for _, f := range files {
					m.Logger.DebugC(fmt.Sprintf(":candy:%s -> %s", k, f))
				}
			}
		}

		// Bump packages
		err = m.BumpAtoms(mkit, opts)
		if err != nil {
			return err
		}

	} else {
		m.Logger.Info(
			":smiling_face_with_sunglasses:No candidates found. Nothing to do.")
	}

	// Copy eclasses
	err = m.MergeEclasses(mkit, opts)
	if err != nil {
		return err
	}

	// Prepare metadata directory of the kit
	err = m.prepareMetadataDir(mkit, opts)
	if err != nil {
		return err
	}

	// Prepare profiles directory of the kit
	err = m.prepareProfilesDir(mkit, candidates, opts)
	if err != nil {
		return err
	}

	// Copy fixups
	err = m.MergeFixups(mkit, opts)
	if err != nil {
		return err
	}

	if opts.Push && m.hasCommit {

		kitDir := filepath.Join(m.GetTargetDir(), targetKit.Name)
		pushOpts := NewPushOptions()
		err = Push(kitDir, pushOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MergeBot) SearchAtoms(mkit *specs.MergeKit, opts *MergeBotOpts) ([]*specs.RepoScanAtom, error) {
	ans := []*specs.RepoScanAtom{}
	for _, atom := range mkit.Target.Atoms {
		m.Logger.InfoC(fmt.Sprintf(":lollipop:[%s] Checking...",
			atom.Package))

		candidate, err := m.searchAtom(atom, mkit, opts)
		if err != nil {
			return ans, err
		}

		if candidate != nil {
			ans = append(ans, candidate)
		}
	}

	return ans, nil
}

func (m *MergeBot) searchAtom(atom *specs.MergeKitAtom, mkit *specs.MergeKit,
	opts *MergeBotOpts) (*specs.RepoScanAtom, error) {

	pOpts := NewPortageResolverOpts()
	pOpts.Conditions = atom.Conditions

	// Retrieve the last version package that matches the
	// conditions.
	ans, err := m.Resolver.GetLastPackage(atom.Package, pOpts)
	if err != nil {
		return ans, err
	}

	if ans == nil {
		// POST: no packages found for atom.
		return nil, nil
	}

	// Check if the selected package is with slot
	gatom, err := gentoo.ParsePackageStr(atom.Package)
	if err != nil {
		return ans, err
	}

	gpkg, err := ans.ToGentooPackage()
	if err != nil {
		return ans, err
	}

	toAdd := true
	existingAtoms, _ := m.TargetResolver.GetPackageVersions(ans.CatPkg)
	if len(existingAtoms) > 0 {
		for _, a := range existingAtoms {
			epkg, err := a.ToGentooPackage()
			if err != nil {
				return ans, err
			}

			// Ignore packages with different SLOTs
			if gatom.Slot != "" && gatom.Slot != "0" && gatom.Slot != epkg.Slot {
				continue
			}

			equal, err := epkg.Equal(gpkg)
			if err != nil {
				return ans, err
			}

			if equal {
				// POST: The package is the same. Checking md5
				if a.Md5 == ans.Md5 {
					toAdd = false
					break
				}
			} else {

				lessThen, err := gpkg.LessThan(epkg)
				if err != nil {
					return ans, err
				}

				if lessThen {

					// TODO: Compare major revision with source.
					//       We can have a major revision of the
					//       same package as autobump logic.
					toAdd = false
					break
				}

			}

		}

	} // else is a new package to add

	if !toAdd {
		ans = nil
	}

	return ans, nil
}

func (m *MergeBot) cloneSourcesKits(mkit *specs.MergeKit, opts *MergeBotOpts) error {
	gitOpts := &CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			SingleBranch: true,
			RemoteName:   "origin",
			Depth:        opts.GitDeepFetch,
		},
		Verbose: true,
		// Always generate summary report
		Summary: true,
		Results: []*specs.ReposcanKit{},
	}

	analysis := &specs.ReposcanAnalysis{
		Kits: mkit.Sources,
	}

	// Clone sources kit
	err := CloneKits(analysis, m.GetSourcesDir(), gitOpts)
	if err != nil {
		return err
	}

	return nil
}

func (m *MergeBot) GenerateReposcanFiles(mkit *specs.MergeKit, opts *MergeBotOpts) error {
	err := helpers.EnsureDirWithoutIds(m.GetReposcanDir(), 0755)
	if err != nil {
		return err
	}

	// Prepare eclass dir list
	eclassDirs := []string{}
	for _, source := range mkit.Sources {
		eclassDir, err := filepath.Abs(filepath.Join(m.GetSourcesDir(), source.Name, "eclass"))
		if err != nil {
			return err
		}
		if utils.Exists(eclassDir) {
			kitDir, _ := filepath.Abs(filepath.Join(m.GetSourcesDir(), source.Name))
			eclassDirs = append(eclassDirs, kitDir)
		}
	}

	for _, source := range mkit.Sources {
		sourceDir := filepath.Join(m.GetSourcesDir(), source.Name)
		targetFile := filepath.Join(m.GetReposcanDir(), source.Name+"-"+source.Branch)

		err := m.GenerateKitCacheFile(sourceDir, source.Name, source.Branch,
			targetFile, eclassDirs, opts.Concurrency)
		if err != nil {
			return err
		}
	}

	// Generate target kit reposcan
	if !m.IsANewBranch {
		kit, _ := mkit.GetTargetKit()
		sourceDir := filepath.Join(m.GetTargetDir(), kit.Name)
		targetFile := filepath.Join(m.GetReposcanDir(), "target-"+kit.Name+"-"+kit.Branch)
		err = m.GenerateKitCacheFile(sourceDir, kit.Name, kit.Branch,
			targetFile, eclassDirs, opts.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MergeBot) cloneTargetKit(mkit *specs.MergeKit, opts *MergeBotOpts) error {
	kit, err := mkit.GetTargetKit()
	if err != nil {
		return err
	}

	m.Logger.Info(fmt.Sprintf(":factory:[%s] Cloning target kit...", kit.Name))
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)

	// Check if the repository branch exists.
	existsBranch, err := BranchExists(kit.Url, kit.Branch)
	if err != nil {
		return err
	}

	gitOpts := &CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			RemoteName: "origin",
			Depth:      opts.GitDeepFetch,
		},
		Verbose: true,
		// Always generate summary report
		Summary: true,
		Results: []*specs.ReposcanKit{},

		SignatureName:  opts.SignatureName,
		SignatureEmail: opts.SignatureEmail,
	}

	if !existsBranch {
		m.Logger.InfoC(fmt.Sprintf("Branch %s doesn't exists. Creating the branch.",
			kit.Branch))

		m.IsANewBranch = true

		err = CloneAndCreateBranch(kit, kitDir, gitOpts)
	} else {
		gitOpts.GitCloneOptions.SingleBranch = true

		err = Clone(kit, kitDir, gitOpts)
		if err != nil {
			return err
		}
	}

	return err
}
