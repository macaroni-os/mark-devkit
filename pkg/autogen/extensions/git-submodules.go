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
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"
)

type ExtensionGitSubmodules struct {
	*ExtensionBase
}

func NewExtensionGitSubmodules(opts map[string]string) (*ExtensionGitSubmodules, error) {
	return &ExtensionGitSubmodules{
		ExtensionBase: &ExtensionBase{
			Opts: opts,
		}}, nil
}

func (e *ExtensionGitSubmodules) GetName() string { return specs.ExtensionGitSubmodules }

func (e *ExtensionGitSubmodules) GetPkgBasedir(mapref *map[string]interface{}) (string, error) {
	values := *mapref

	dir, present := values["pkg_basedir"].(string)
	if !present {
		return "", fmt.Errorf("variable pkg_basedir not present")
	}

	ans, err := helpers.RenderContentWithTemplates(
		dir,
		"", "", "pkg_basedir", values, []string{},
	)

	return ans, err
}

func (e *ExtensionGitSubmodules) Elaborate(restGuard *guard.RestGuard,
	atom, def *specs.AutogenAtom,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref

	downloadDir := e.Opts["download_dir"]
	if !filepath.IsAbs(downloadDir) {
		downloadDir, _ = filepath.Abs(downloadDir)
	}
	// Create the temporary directory of the package
	workdir, _ := e.Opts["workdir"]
	if !filepath.IsAbs(workdir) {
		workdir, _ = filepath.Abs(workdir)
	}
	version, _ := values["version"].(string)

	cloneDir := filepath.Join(workdir, "gitsubmodules-extension", atom.Name)

	err := os.MkdirAll(cloneDir, os.ModePerm)
	if err != nil {
		return err
	}

	mirror := e.Opts["mirror"]
	if mirror == "" {
		mirror = "mirror://macaroni"
	}
	bundleIdentifier := e.Opts["bundle_identifier"]
	if bundleIdentifier == "" {
		bundleIdentifier = "mark-gitsubmodules-bundle"
	}
	bundleExtension := e.Opts["bundle_extension"]
	if bundleExtension == "" {
		bundleExtension = "xz"
	}
	// Fix typo on bundle_extension value.
	bundleExtension = strings.ReplaceAll(bundleExtension, ".", "")
	values["mirror"] = mirror

	// Retrieve tag information
	// NOTE: At the moment only github generator is supported

	sha, present := values["sha"].(string)
	if !present {
		return fmt.Errorf("No sha value found")
	}
	baseUnpackDir, err := e.GetPkgBasedir(mapref)
	if err != nil {
		return err
	}

	pkgWorkDir := filepath.Join(cloneDir, baseUnpackDir)

	gitRepo, present := values["git_repo"].(string)

	cloneOpts := &kit.CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			RemoteName:        "origin",
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		},
		Verbose: log.Config.GetGeneral().Debug,
		Summary: false,
		Results: []*specs.ReposcanKit{},
	}

	if e.Opts["git_single_branch"] == "true" {
		cloneOpts.GitCloneOptions.SingleBranch = true
	}

	if e.Opts["git_depth"] != "" {
		d, _ := strconv.Atoi(e.Opts["git_depth"])
		if d > 0 {
			cloneOpts.GitCloneOptions.Depth = d
		}
	}

	repoData := &specs.ReposcanKit{
		Name:       atom.Name,
		Url:        gitRepo,
		Branch:     "",
		CommitSha1: sha,
	}

	log.DebugC(fmt.Sprintf(":factory:[%s] Cloning repo %s and fetch submodules...",
		atom.Name, gitRepo))

	// Clone repo and submodules.
	err = kit.Clone(repoData, pkgWorkDir, cloneOpts)
	if err != nil {
		return err
	}

	if len(cloneOpts.Submodules) == 0 {
		return fmt.Errorf("package %s without submodules!", atom.Name)
	}

	bundleTarball := filepath.Join(
		downloadDir, fmt.Sprintf(
			"%s-%s-%s-%s.tar.%s", atom.Name, version, bundleIdentifier,
			sha[0:7], bundleExtension,
		))

	// Create bundle tarball
	archiveDirs := []string{}
	for _, sub := range cloneOpts.Submodules {
		subStatus, _ := sub.Status()
		archiveDirs = append(archiveDirs, filepath.Join(pkgWorkDir, subStatus.Path))
	}
	bundleArt, err := e.createBundleTarball(
		bundleTarball, cloneDir, atom.Name, mirror, archiveDirs)
	if err != nil {
		return err
	}

	// Retrieve artefacts.
	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)

	artefacts = append(artefacts, bundleArt)
	values["artefacts"] = artefacts

	e.cleanup(mapref)

	if !logger.GetDefaultLogger().Config.GetGeneral().Debug {
		defer os.RemoveAll(filepath.Join(workdir, "gitsubmodules-extension"))
	}

	return nil
}

func (e *ExtensionGitSubmodules) createBundleTarball(
	bundleTarball, cloneDir, atomName, mirror string,
	archiveDirs []string) (*specs.AutogenArtefact, error) {

	var err error
	log := logger.GetDefaultLogger()

	tarformers := executor.NewTarFormers(e.getTarformersConfig())
	s := tarf_specs.NewSpecFile()
	// We don't need to keep the original permission of the files
	// and owner.
	s.SameOwner = false
	s.SameChtimes = false
	s.Writer = tarf_specs.NewWriter()
	s.Writer.ArchiveDirs = archiveDirs

	s.RenamePath = []tarf_specs.RenameRule{
		{
			Source: cloneDir,
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
