/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	autogenart "github.com/macaroni-os/mark-devkit/pkg/autogen/artefacts"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"github.com/pelletier/go-toml/v2"
)

type ExtensionRust struct {
	*ExtensionBase
}

type CargoLock struct {
	Version  int            `toml:"version,omitempty"`
	Packages []CargoPackage `toml:"package"`
}

type CargoWorkspace struct {
	Package      *CargoPackage `toml:"package,omitempty"`
	Members      []string      `toml:"members,omitempty"`
	Exclude      []string      `toml:"exclude,omitempty"`
	Dependencies []string      `toml:"dependencies,omitempty"`
}

type CargoToml struct {
	Package   *CargoPackage   `toml:"package,omitempty"`
	Workspace *CargoWorkspace `toml:"workspace,omitempty"`
	Path      string
}

type CargoPackage struct {
	Name         string   `toml:"name,omitempty"`
	Version      any      `toml:"version,omitempty"`
	Source       any      `toml:"source,omitempty"`
	Checksum     string   `toml:"checksum,omitempty"`
	Dependencies []string `toml:"dependencies,omitempty"`
}

type CargoLocalDeps struct {
	Map map[string]*CargoToml
}

func (cp *CargoPackage) GetVersion() string {
	// We need to use any in order to avoid
	// exception if a specific Cargo.toml set version
	// or other attributes as string.
	// For example this break string conversion:
	// version.workspace = true
	if version, ok := cp.Version.(string); ok {
		return version
	}
	return ""
}

func (cp *CargoPackage) GetSource() string {
	if source, ok := cp.Version.(string); ok {
		return source
	}
	return ""
}

func NewExtensionRust(opts map[string]string) (*ExtensionRust, error) {
	return &ExtensionRust{
		ExtensionBase: &ExtensionBase{
			Opts: opts,
		}}, nil
}

func (e *ExtensionRust) GetName() string { return specs.ExtensionRust }

func (e *ExtensionRust) GetCratesSkipped() map[string]bool {
	ans := make(map[string]bool)

	crates, _ := e.Opts["crates_skipped"]
	if crates != "" {
		for _, c := range strings.Split(crates, " ") {
			ans[strings.TrimSpace(c)] = true
		}
	}

	return ans
}

func (e *ExtensionRust) Elaborate(restGuard *guard.RestGuard,
	atom, def *specs.AutogenAtom,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref

	// Create the temporary directory of the package
	workdir, _ := e.Opts["workdir"]
	downloadDir := e.Opts["download_dir"]
	if !filepath.IsAbs(downloadDir) {
		downloadDir, _ = filepath.Abs(downloadDir)
	}
	if !filepath.IsAbs(workdir) {
		workdir, _ = filepath.Abs(workdir)
	}

	pkgWorkDir := filepath.Join(workdir, "rust-extension", atom.Name)

	err := os.MkdirAll(pkgWorkDir, os.ModePerm)
	if err != nil {
		return err
	}

	mirror := e.Opts["mirror"]
	if mirror == "" {
		mirror = "mirror://macaroni"
	}
	bundleIdentifier := e.Opts["bundle_identifier"]
	if bundleIdentifier == "" {
		bundleIdentifier = "mark-rust-bundle"
	}
	bundleExtension := e.Opts["bundle_extension"]
	if bundleExtension == "" {
		bundleExtension = "xz"
	}
	// Fix typo on bundle_extension value.
	bundleExtension = strings.ReplaceAll(bundleExtension, ".", "")
	values["mirror"] = mirror

	// Ensure download dir. Could be not present the first time.
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Retrieve artefacts.
	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)
	// NOTE: I consider the first artefact the main for golang vendor generation.
	if len(artefacts) == 0 {
		log.DebugC(
			fmt.Sprintf(
				"[%s] No artefacts found for rust vendor generation. Nothing to do.",
				atom.Name,
			))
		return nil
	}

	art := artefacts[0]

	// Download the main artefacts to the download dir.
	log.DebugC(
		fmt.Sprintf("[%s] Downloading %s from url %s",
			atom.Name, art.Name, art.SrcUri[0],
		))
	repoFile, err := autogenart.DownloadArtefact(restGuard, atom,
		art.SrcUri[0], art.Name, downloadDir)
	if err != nil {
		return err
	}

	// Unpack the tarballs
	unpackDir := filepath.Join(pkgWorkDir, "unpack")
	log.Info(
		fmt.Sprintf(":factory:[%s] Extracting file %s",
			atom.Name, repoFile.Name,
		))
	err = e.unpackArtefact(downloadDir, unpackDir, repoFile,
		atom, def, mapref)
	if err != nil {
		return err
	}

	// Retrieve Cargo.lock
	pkgUnpackDir, cargoLock, err := e.retrieveCargoLock(atom, unpackDir)
	if err != nil {
		return err
	}
	values["pkg_basedir"] = filepath.Base(pkgUnpackDir)

	// Read all Cargo.toml in order to identify all local crates
	localCrates, err := e.parseCargoToml(atom, pkgUnpackDir, true)
	if err != nil {
		return fmt.Errorf(
			"Error on parse cargo toml: %s", err.Error())
	}

	// Download cargo bundles files
	bundlesDir := filepath.Join(pkgWorkDir, bundleIdentifier+"-"+atom.Name)
	err = e.downloadBundles(restGuard, atom, cargoLock, localCrates,
		pkgUnpackDir, bundlesDir)
	if err != nil {
		return err
	}

	// Create bundle tarball
	version, _ := values["version"].(string)
	sha, _ := values["sha"].(string)
	bundleTarball := ""
	if sha == "" {
		bundleTarball = fmt.Sprintf(
			"%s-%s-%s.tar.%s", atom.Name, version, bundleIdentifier, bundleExtension,
		)
	} else {
		bundleTarball = fmt.Sprintf(
			"%s-%s-%s-%s.tar.%s", atom.Name, version, bundleIdentifier,
			sha[0:7], bundleExtension,
		)
	}
	bundleArt, err := e.createBundleTarball(bundlesDir,
		filepath.Join(downloadDir, bundleTarball), workdir, atom.Name, mirror)
	if err != nil {
		return err
	}

	artefacts = append(artefacts, bundleArt)
	values["artefacts"] = artefacts

	e.cleanup(mapref)

	if !logger.GetDefaultLogger().Config.GetGeneral().Debug {
		defer os.RemoveAll(filepath.Join(workdir, "rust-extension"))
	}

	return nil
}

func (e *ExtensionRust) parseCargoToml(atom *specs.AutogenAtom,
	pkgUnpackDir string, excludeMainCargo bool) (*CargoLocalDeps, error) {

	var err error
	ans := &CargoLocalDeps{
		Map: make(map[string]*CargoToml, 0),
	}

	// file to skip
	toSkip := ""
	if excludeMainCargo {
		toSkip = filepath.Join(pkgUnpackDir, "Cargo.toml")
	}

	err = e.checkDir4CargoTomp(pkgUnpackDir, toSkip, ans)

	return ans, err
}

func (e *ExtensionRust) createBundleTarball(
	bundlesDir, bundleTarball, workDir, atomName, mirror string) (*specs.AutogenArtefact, error) {
	var err error
	log := logger.GetDefaultLogger()

	tarformers := executor.NewTarFormers(e.getTarformersConfig())
	s := tarf_specs.NewSpecFile()
	// We don't need to keep the original permission of the files
	// and owner.
	s.SameOwner = false
	s.SameChtimes = false
	s.Writer = tarf_specs.NewWriter()
	s.Writer.ArchiveDirs = []string{bundlesDir}

	s.RenamePath = []tarf_specs.RenameRule{
		{
			Source: filepath.Join(workDir, "rust-extension", atomName),
			Dest:   ".",
		},
	}
	tarfOpts := tools.NewTarCompressionOpts(true)
	defer tarfOpts.Close()

	err = tools.PrepareTarWriter(bundleTarball, tarfOpts)
	if err != nil {
		return nil, fmt.Errorf(
			"error on prepare writer: %s",
			err.Error())
	}

	if tarfOpts.CompressWriter != nil {
		tarformers.SetWriter(tarfOpts.CompressWriter)
	} else {
		tarformers.SetWriter(tarfOpts.FileWriter)
	}

	fileName := filepath.Base(bundleTarball)
	log.Debug(fmt.Sprintf("[%s] Creating tarball %s...",
		atomName, fileName))

	err = tarformers.RunTaskWriter(s)
	if err != nil {
		return nil, fmt.Errorf(
			"error on create tarball %s: %s",
			fileName, err.Error())
	}

	artUri := fmt.Sprintf("%s/%s", mirror, fileName)

	local := true
	return &specs.AutogenArtefact{
		SrcUri: []string{artUri},
		Name:   fileName,
		Local:  &local,
	}, nil
}

func (e *ExtensionRust) downloadBundles(restGuard *guard.RestGuard,
	atom *specs.AutogenAtom, cargoLock *CargoLock, localCrates *CargoLocalDeps,
	unpackDir, bundlesDir string) error {
	log := logger.GetDefaultLogger()
	var err error
	var sourceOrigin string

	// Create bundle dir
	err = os.MkdirAll(bundlesDir, os.ModePerm)
	if err != nil {
		return err
	}

	skipMap := e.GetCratesSkipped()

	// NOTE: Git crates could reference multiple
	// packages with the same url. So,
	// we need generate the tarball only
	// after the processing of all packages.
	gitDeps := make(map[string][]*CargoPackage, 0)

	for _, pkg := range cargoLock.Packages {

		if strings.HasPrefix(pkg.GetSource(), "git+") {
			sourceOrigin = "git"
		} else {
			sourceOrigin = "crates"
		}

		if sourceOrigin == "crates" {

			if _, skipped := skipMap[pkg.Name]; skipped {
				log.Debug(fmt.Sprintf(
					"[%s] Ignored bundle %s %s skipped.",
					atom.Name, pkg.Name, pkg.Version))
				continue
			}

			if _, local := localCrates.Map[pkg.Name]; local {
				log.Debug(fmt.Sprintf(
					"[%s] Ignored local bundle %s %s skipped.",
					atom.Name, pkg.Name, pkg.Version))
				continue
			}

			url := fmt.Sprintf(
				"https://crates.io/api/v1/crates/%s/%s/download",
				pkg.Name, pkg.Version,
			)

			// Define the name of the bundle to download
			bundle := fmt.Sprintf("%s-%s.crate", pkg.Name, pkg.Version)

			log.Debug(fmt.Sprintf("[%s] Downloading bundle %s %s at %s...",
				atom.Name, pkg.Name, pkg.Version, url))

			_, err = autogenart.DownloadArtefact(
				restGuard, atom, url,
				bundle, bundlesDir)
			if err != nil {
				return err
			}

		} else {
			// Keep old logic for now. Maybe we can avoid this.
			// Drop initial git+ string
			if pkg.GetSource() != "" {
				url := pkg.GetSource()[4:]
				gitDepPkgs, present := gitDeps[url]
				if present {
					gitDepPkgs = append(gitDepPkgs, &pkg)
				} else {
					gitDepPkgs = []*CargoPackage{&pkg}
				}
				gitDeps[url] = gitDepPkgs
			}
		}

	}

	if len(gitDeps) > 0 {
		for url, gitPkgs := range gitDeps {
			// POST: git bundle
			err = e.processGitCrate(atom, gitPkgs, url, bundlesDir, unpackDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *ExtensionRust) processGitCrate(atom *specs.AutogenAtom,
	pkgs []*CargoPackage, url, bundlesDir, unpackDir string) error {
	log := logger.GetDefaultLogger()

	ref := strings.Split(url, "#")[1]
	gitRepo := strings.Split(url, "?")[0]

	// Name of the dir of the bundle unpacked.
	// We escape chars: [:] -> %3A, [/] -> %2F
	cloneBasenameDir := strings.ReplaceAll(
		strings.ReplaceAll(gitRepo, ":", "%3A"),
		"/", "%2F") + "-" + ref
	cloneDir := filepath.Join(unpackDir, "staging", cloneBasenameDir)

	singleBranch := true
	if e.Opts["crates_git_single_branch"] == "false" {
		singleBranch = false
	}

	cloneOpts := &kit.CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			SingleBranch: singleBranch,
			RemoteName:   "origin",
			//Depth:        10000,
			// Enable submodule cloning.
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		},
		Verbose: log.Config.GetGeneral().Debug,
		Summary: false,
		Results: []*specs.ReposcanKit{},
	}

	if e.Opts["crates_git_depth"] != "" {
		d, _ := strconv.Atoi(e.Opts["crates_git_depth"])
		if d > 0 {
			cloneOpts.GitCloneOptions.Depth = d
		}
	}

	// Create ReposcanKit with the git repo data
	repoData := &specs.ReposcanKit{
		Name:       pkgs[0].Name,
		Url:        gitRepo,
		Branch:     "",
		CommitSha1: ref,
	}

	// Clone repo and submodules.
	err := kit.Clone(repoData, cloneDir, cloneOpts)
	if err != nil {
		return err
	}

	// Retrieve the local creates of the unpack dir
	gitRepoCrates, err := e.parseCargoToml(atom, cloneDir, false)
	if err != nil {
		return err
	}

	// Generate the mark_config.toml file under the unpack directory.
	markConfigContent := "[patch.'" + gitRepo + "']\n"
	for _, pkg := range pkgs {

		// Retrieve crate path
		localCrate, available := gitRepoCrates.Map[pkg.Name]
		if !available {
			return fmt.Errorf("crate %s not found on git repo %s",
				pkg.Name, gitRepo)
		}

		if cloneDir == localCrate.Path {
			markConfigContent += pkg.Name +
				" = { path = \"%CRATES_DIR%/" + cloneBasenameDir + "\" }\n"
		} else {
			rel := localCrate.Path[len(cloneDir)+1:]
			markConfigContent += pkg.Name +
				" = { path = \"%CRATES_DIR%/" + cloneBasenameDir + "/" + rel + "\" }\n"
		}
	}

	err = os.WriteFile(
		filepath.Join(cloneDir, "mark_config.toml"),
		[]byte(markConfigContent), 0644)
	if err != nil {
		return err
	}

	// Create bundle of the git crate
	archiveName := fmt.Sprintf("%s.tar.xz", cloneBasenameDir)
	tarformers := executor.NewTarFormers(e.getTarformersConfig())
	s := tarf_specs.NewSpecFile()
	// We don't need to keep the original permission of the files
	// and owner.
	s.SameOwner = false
	s.SameChtimes = false
	s.Writer = tarf_specs.NewWriter()
	s.Writer.ArchiveDirs = []string{cloneDir}

	s.RenamePath = []tarf_specs.RenameRule{
		{
			Source: filepath.Join(unpackDir, "staging"),
			Dest:   ".",
		},
	}
	tarfOpts := tools.NewTarCompressionOpts(true)
	defer tarfOpts.Close()

	err = tools.PrepareTarWriter(
		filepath.Join(bundlesDir, archiveName), tarfOpts)
	if err != nil {
		return fmt.Errorf(
			"error on prepare writer: %s",
			err.Error())
	}

	if tarfOpts.CompressWriter != nil {
		tarformers.SetWriter(tarfOpts.CompressWriter)
	} else {
		tarformers.SetWriter(tarfOpts.FileWriter)
	}

	log.Debug(fmt.Sprintf("[%s] Creating tarball %s...",
		atom.Name, archiveName))

	err = tarformers.RunTaskWriter(s)
	if err != nil {
		return fmt.Errorf(
			"error on create tarball %s: %s",
			archiveName, err.Error())
	}

	return nil
}

func (e *ExtensionRust) retrieveCargoLock(atom *specs.AutogenAtom,
	targetDir string) (string, *CargoLock, error) {

	pkgUnpackDir := ""
	unpackDirPrefix, _ := e.Opts["unpack_srcdir_prefix"]
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return "", nil, err
	}

	cargoLockPath := ""

	for _, entry := range entries {

		if !entry.IsDir() {
			continue
		}

		pkgUnpackDir = filepath.Join(targetDir, entry.Name())
		cargoLockPath = filepath.Join(pkgUnpackDir, "Cargo.lock")
		if unpackDirPrefix != "" && strings.HasPrefix(entry.Name(), unpackDirPrefix) {
			if !utils.Exists(cargoLockPath) {
				cargoLockPath = ""
				break
			}
		}

		if utils.Exists(cargoLockPath) {
			break
		}
		cargoLockPath = ""
	}

	if cargoLockPath == "" {
		return "", nil, fmt.Errorf("Cargo.lock file not found")
	}

	data, err := os.ReadFile(cargoLockPath)
	if err != nil {
		return "", nil, fmt.Errorf("error on read file Cargo.lock: %s", err.Error())
	}

	ans := &CargoLock{}
	err = toml.Unmarshal([]byte(data), ans)
	if err != nil {
		return "", nil, fmt.Errorf("error on unmarshal Cargo.lock: %s", err.Error())
	}

	return pkgUnpackDir, ans, nil
}

func (e *ExtensionRust) checkDir4CargoTomp(dir, toSkip string,
	deps *CargoLocalDeps) error {

	cfile := filepath.Join(dir, "Cargo.toml")
	if utils.Exists(cfile) {
		if cfile != toSkip {
			data, err := os.ReadFile(cfile)
			if err != nil {
				return fmt.Errorf("error on read file %s: %s",
					cfile, err.Error())
			}

			pkg := CargoToml{Path: dir}
			err = toml.Unmarshal([]byte(data), &pkg)
			if err != nil {
				if derr, ok := err.(*toml.DecodeError); ok {
					row, col := derr.Position()
					return fmt.Errorf("error on parse cargo toml %s (row %d, column %d): %s",
						cfile, row, col, derr.Error())
				}
				return fmt.Errorf("error on unmarshal file %s: %s",
					cfile, err)
			}

			if pkg.Package == nil && pkg.Workspace == nil {
				return fmt.Errorf("unexpected file content on %s",
					cfile)
			}

			if pkg.Package == nil {
				pkg.Package = pkg.Workspace.Package
			}

			if pkg.Package.Name != "" {
				deps.Map[pkg.Package.Name] = &pkg
			}
		}
	}

	fEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range fEntries {
		if entry.IsDir() {
			err := e.checkDir4CargoTomp(
				filepath.Join(dir, entry.Name()), toSkip, deps)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
