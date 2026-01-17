/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	autogenart "github.com/macaroni-os/mark-devkit/pkg/autogen/artefacts"
	tmpleng "github.com/macaroni-os/mark-devkit/pkg/autogen/tmpl-engines"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (a *AutogenBot) GeneratePackageOnStaging(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom,
	mapref *map[string]interface{},
	tmplEngine tmpleng.TemplateEngine) (*specs.RepoScanAtom, error) {

	values := *mapref
	category := atom.GetCategory(def)
	targetKit, _ := mkit.GetTargetKit()
	pkgDirStaging := filepath.Join(a.GetSourcesDir(), targetKit.Name,
		category, atom.Name)

	pn, _ := values["pn"].(string)
	version, _ := values["version"].(string)
	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)
	slot := helpers.GetSlotFromValues(mapref)

	ans := &specs.RepoScanAtom{
		Atom:     fmt.Sprintf("%s/%s-%s", category, atom.Name, version),
		Category: category,
		Package:  pn,
		Revision: "0",
		CatPkg:   fmt.Sprintf("%s/%s", category, atom.Name),
		Kit:      targetKit.Name,
		Branch:   targetKit.Branch,
		Files:    []specs.RepoScanFile{},
		// We need to have KEYWORDS for merge check
		Metadata: map[string]string{
			"KEYWORDS": "*",
		},
	}

	if slot != "" && slot != "0" {
		ans.Metadata["SLOT"] = slot
	} else {
		ans.Metadata["SLOT"] = "0"
	}

	// Prepare package dir
	err := helpers.EnsureDirWithoutIds(pkgDirStaging, 0755)
	if err != nil {
		return nil, err
	}

	// Special vars rendered
	renderedVars := []string{
		"body", "iuse", "rdepend", "bdepend", "depend", "pdepend",
		"cdepend", "s", "homepage", "desc", "required_use",
	}
	for _, field := range renderedVars {
		if _, hasField := values[field]; hasField {
			// Render the body with the values.
			fieldValue, _ := values[field].(string)

			values[field], err = helpers.RenderContentWithTemplates(
				fieldValue,
				"", "", "ebuild."+field, values, []string{},
			)
			if err != nil {
				a.Logger.Warning(fmt.Sprintf("[%s] Error on render variable %s: %s",
					atom.Name, field, err.Error()))
			}
		}
	}

	if len(artefacts) > 0 && (atom.IgnoreArtefacts == nil || !*atom.IgnoreArtefacts) {
		// Download tarballs
		for idx, art := range artefacts {

			var repoFile *specs.RepoScanFile
			var err error

			filename := art.Name
			if art.Name == "" {
				filename = fmt.Sprintf("%s-%s.tar.gz", pn, version)
			}

			if art.Local != nil && *art.Local {

				a.Logger.DebugC(
					fmt.Sprintf(":factory: [%s] Elaborating local file %s with url %s",
						atom.Name, art.Name, art.SrcUri[0],
					))

				repoFile, err = a.processLocalArtefact(atom, art.SrcUri[0], art.Name)
			} else {

				a.Logger.DebugC(
					fmt.Sprintf("[%s] Downloading %s from url %s",
						atom.Name, art.Name, art.SrcUri[0],
					))

				repoFile, err = autogenart.DownloadArtefact(
					a.RestGuard, atom, art.SrcUri[0], art.Name, a.GetDownloadDir())
			}

			if err != nil {
				return nil, fmt.Errorf("On downloading %s from url %s: %s",
					art.Name, art.SrcUri[0], err,
				)
			}

			ans.Files = append(ans.Files, *repoFile)

			if idx == 0 {
				// Set the first artefact as src_uri. This variable
				// could be used to avoid iteration of the artefacts
				// in the template when we are sure that there is
				// only one artefacts or the url of the first artefact
				// is used somewhere.
				values["src_uri"] = fmt.Sprintf("%s -> %s", art.SrcUri[0], filename)
			}

		}

		manifestPath := filepath.Join(pkgDirStaging, "Manifest")
		var manifest *kit.ManifestFile
		// Check if exists already a Manifest with other files.
		// For example the same package is been defined multiple time
		// with multiple selector.
		if utils.Exists(manifestPath) {
			manifest, err = kit.ParseManifest(manifestPath)
			if err != nil {
				return nil, err
			}
			manifest.AddFiles(ans.Files)
		} else {
			manifest = kit.NewManifestFile(ans.Files)
		}

		// Generate Manifest
		err := manifest.Write(manifestPath)
		if err != nil {
			return nil, err
		}
	}

	ebuildPath := filepath.Join(pkgDirStaging, fmt.Sprintf("%s-%s.ebuild", pn, version))
	err = tmplEngine.Render(aspec, atom, def, mapref, ebuildPath)
	if err != nil {
		return ans, err
	}

	filesDirPath := filepath.Join(filepath.Dir(aspec.File), atom.Category, pn, "files")

	// Process files dir
	if atom.FilesDir != "" {
		filesDirPath, err = helpers.RenderContentWithTemplates(
			atom.FilesDir,
			"", "", "atom.filesdir", values, []string{},
		)
		if err != nil {
			return ans, err
		}
		// Always create the path based on spec file path as base
		filesDirPath = filepath.Join(filepath.Dir(aspec.File), filesDirPath)
	}

	if utils.Exists(filesDirPath) {
		pkgFilesDir := filepath.Join(pkgDirStaging, "files")
		err = a.copyFilesDir(filesDirPath, pkgFilesDir)
	}

	return ans, err
}

func (a *AutogenBot) processLocalArtefact(atom *specs.AutogenAtom,
	atomUrl, tarballName string) (*specs.RepoScanFile, error) {

	ans := &specs.RepoScanFile{
		SrcUri: []string{atomUrl},
		Name:   tarballName,
		Hashes: make(map[string]string, 0),
	}

	artefactPath := filepath.Join(a.GetDownloadDir(), tarballName)

	hashes, err := helpers.GetFileHashes(artefactPath)
	if err != nil {
		return nil, err
	}

	ans.Hashes["sha512"] = hashes.Sha512()
	ans.Hashes["blake2b"] = hashes.Blake2b()
	ans.Size = fmt.Sprintf("%d", hashes.Size())

	return ans, nil
}

func (a *AutogenBot) copyFilesDir(sourcedir, targetdir string) error {

	a.Logger.Debug(fmt.Sprintf("Analyzing directory %s...", sourcedir))

	entries, err := os.ReadDir(sourcedir)
	if err != nil {
		return fmt.Errorf("error on reading dir %s: %s",
			sourcedir, err.Error())
	}

	if !utils.Exists(targetdir) {
		err = helpers.EnsureDirWithoutIds(targetdir, 0755)
		if err != nil {
			return fmt.Errorf(
				"error on create dir %s: %s",
				targetdir, err.Error())
		}
	}

	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := a.copyFilesDir(
				filepath.Join(sourcedir, entry.Name()),
				filepath.Join(targetdir, entry.Name()),
			)
			if err != nil {
				return fmt.Errorf(
					"error on copy subdir %s: %s",
					entry.Name(), err.Error())
			}

		} else {

			sourceFile := filepath.Join(sourcedir, entry.Name())
			targetFile := filepath.Join(targetdir, entry.Name())

			fi, _ := os.Lstat(sourceFile)
			if fi.Mode()&fs.ModeSymlink != 0 {
				// POST: file is a link.

				if utils.Exists(targetFile) {
					// Keep things easy. If the path
					// exists I just avoid to manage all the
					// possible use cases. We can fix the link
					// manually eventually.
					continue
				}

				symlink, err := os.Readlink(sourceFile)
				if err != nil {
					return fmt.Errorf(
						"error on readlink %s: %s",
						sourceFile, err.Error())
				}

				if err = os.Symlink(symlink, targetFile); err != nil {
					return fmt.Errorf(
						"error on create symlink %s -> %s: %s",
						symlink, targetFile, err.Error())
				}

			} else {
				err = helpers.CopyFile(sourceFile, targetFile)
				if err != nil {
					return fmt.Errorf(
						"error on copy file %s -> %s: %s",
						sourceFile, targetFile,
						err.Error())
				}
			}
		}
	}

	return nil
}

func (a *AutogenBot) isVersion2Add(atom, def *specs.AutogenAtom,
	version string, opts *AutogenBotOpts,
	mapref *map[string]interface{}) (bool, error) {

	if a.MergeBot.TargetKitIsANewBranch() {
		return true, nil
	}

	catpkg := fmt.Sprintf("%s/%s", atom.GetCategory(def), atom.Name)

	if !a.MergeBot.TargetResolver.IsPresentPackage(catpkg) {
		a.Logger.Debug(fmt.Sprintf(
			":eyes:[%s] Package is not present on target resolver.", atom.Name))
		return true, nil
	}

	pOpts := kit.NewPortageResolverOpts()

	// When selector are defined we need to retrieve packages matching
	// with the selector
	if atom.HasSelector() {
		pOpts.Conditions = []string{}
		// Normally the versions downloaded from public
		// website are without SLOT and it doesn't make sense
		// try to compare versions with conditions having SLOT.
		pOpts.IgnoreSlot = true

		if atom.HasSelector4Slot() {
			// NOTE: There are particular use case where the same version
			//       is used with different revision and different slots
			//       (for example for webkit-gtk). In this case, we need
			//       to use the slot for compare versions.
			pOpts.IgnoreSlot = !atom.GetSelector4Slot()

			a.Logger.Debug(fmt.Sprintf(
				":eyes:[%s] Using value %v for SLOT in selector.",
				atom.Name, pOpts.IgnoreSlot))
		}

		for _, condition := range atom.Selector {
			gpkgCond, err := helpers.DecodeCondition(condition,
				atom.GetCategory(def), atom.Name,
			)
			if err != nil {
				return false, err
			}
			if pOpts.IgnoreSlot {
				pOpts.Conditions = append(pOpts.Conditions, fmt.Sprintf("%s-%s",
					gpkgCond.GetPackageNameWithCond(), gpkgCond.GetPV()))
			} else {
				slot := helpers.GetSlotFromValues(mapref)

				pOpts.Conditions = append(pOpts.Conditions, fmt.Sprintf("%s-%s:%s",
					gpkgCond.GetPackageNameWithCond(), gpkgCond.GetPV(), slot))
			}
		}

		a.Logger.Debug(fmt.Sprintf(
			":eyes:[%s] Package with conditions %s.", atom.Name, pOpts.Conditions))
	}

	// Retrieve all availables version in order to return true
	// and check for differences.
	atomsAvailables, err := a.MergeBot.TargetResolver.GetValidPackages(catpkg, pOpts)
	if err != nil {
		return false, err
	}

	// Prepare GentooPackage of the selected version
	gpkg, err := gentoo.ParsePackageStr(fmt.Sprintf("%s-%s", catpkg, version))
	if err != nil {
		return false, err
	}

	toAdd := true
	for idx := range atomsAvailables {
		agpg, _ := atomsAvailables[idx].ToGentooPackage()
		if equal, _ := agpg.Equal(gpkg); equal {
			if !opts.MergeForced {
				toAdd = false
				break
			}
		}

		if toskip, _ := agpg.GreaterThan(gpkg); toskip {
			toAdd = false
		}

	}

	return toAdd, nil
}
