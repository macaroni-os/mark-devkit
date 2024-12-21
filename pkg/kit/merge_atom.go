/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (m *MergeBot) MergeAtoms(candidates []*specs.RepoScanAtom,
	mkit *specs.MergeKit, opts *MergeBotOpts) error {

	for _, atom := range candidates {
		err := m.mergeAtom(atom, mkit, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MergeBot) mergeAtom(atom *specs.RepoScanAtom,
	mkit *specs.MergeKit, opts *MergeBotOpts) error {
	var ebuildFile, manifestFile string
	kit, _ := mkit.GetTargetKit()

	m.Logger.Debug(fmt.Sprintf(
		"[%s] merging...", atom.Atom))

	targetPkgDir := filepath.Join(m.GetTargetDir(), kit.Name,
		atom.Category, atom.Package)
	sourcePkgDir := filepath.Join(m.GetSourcesDir(), atom.Kit,
		atom.Category, atom.Package)

	if !utils.Exists(sourcePkgDir) {
		return fmt.Errorf("Atom %s not found on kit %s!",
			atom.Atom, atom.Kit)
	}

	if !utils.Exists(targetPkgDir) {
		err := helpers.EnsureDirWithoutIds(targetPkgDir, 0755)
		if err != nil {
			return err
		}
	}

	yamlAtom, _ := atom.Yaml()
	m.Logger.DebugC(fmt.Sprintf(
		"Processing atom %s:\n%s",
		atom.Atom, yamlAtom))

	// Create ebuild
	ebuildFile, err := m.copyEbuild(sourcePkgDir, targetPkgDir, atom)
	if err != nil {
		return err
	}

	// Create manifest
	manifestFile, err = m.createManifest(targetPkgDir, atom)
	if err != nil {
		return fmt.Errorf("error on create manifest for %s: %s",
			atom.Atom, err.Error())
	}

	files4commit := []string{ebuildFile, manifestFile}

	filesDir := filepath.Join(sourcePkgDir, "files")
	if utils.Exists(filesDir) {
		// Add files under <pkgdir>/files directory if there are changes.
		files, err := m.copyFilesDir(filesDir,
			filepath.Join(targetPkgDir, "files"))
		if err != nil {
			return err
		}
		files4commit = append(files4commit, files...)
	}

	m.files4Commit[atom.Atom] = files4commit

	return nil
}

func (m *MergeBot) copyFilesDir(sourcedir, targetdir string) ([]string, error) {
	ans := []string{}

	m.Logger.Debug(fmt.Sprintf("Analyzing directory %s...", sourcedir))

	entries, err := os.ReadDir(sourcedir)
	if err != nil {
		return ans, fmt.Errorf("error on reading dir %s: %s",
			sourcedir, err.Error())
	}

	if !utils.Exists(targetdir) {
		err = helpers.EnsureDirWithoutIds(targetdir, 0755)
		if err != nil {
			return ans, fmt.Errorf(
				"error on create dir %s: %s",
				targetdir, err.Error())
		}
	}

	if len(entries) > 0 {
		for _, entry := range entries {
			if entry.IsDir() {
				files, err := m.copyFilesDir(
					filepath.Join(sourcedir, entry.Name()),
					filepath.Join(targetdir, entry.Name()),
				)
				if err != nil {
					return ans, fmt.Errorf(
						"error on copy subdir %s: %s",
						entry.Name(), err.Error())
				}

				ans = append(ans, files...)
			} else {

				sourceFile := filepath.Join(sourcedir, entry.Name())
				targetFile := filepath.Join(targetdir, entry.Name())
				targetFileMd5 := ""

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
						return ans, fmt.Errorf(
							"error on readlink %s: %s",
							sourceFile, err.Error())
					}

					if err = os.Symlink(symlink, targetFile); err != nil {
						return ans, fmt.Errorf(
							"error on create symlink %s -> %s: %s",
							symlink, targetFile, err.Error())
					}

				} else {
					sourceFileMd5, err := helpers.GetFileMd5(sourceFile)
					if err != nil {
						return ans, fmt.Errorf(
							"error on retrieve md5 of file %s: %s",
							sourceFile, err.Error())
					}

					if utils.Exists(targetFile) {
						targetFileMd5, err = helpers.GetFileMd5(targetFile)
						if err != nil {
							return ans, fmt.Errorf(
								"error on retrieve md5 of file %s: %s",
								targetFile, err.Error())
						}
					}

					if sourceFileMd5 == targetFileMd5 {
						continue
					}

					err = helpers.CopyFile(sourceFile, targetFile)
					if err != nil {
						return ans, fmt.Errorf(
							"error on copy file %s -> %s: %s",
							sourceFile, targetFile,
							err.Error())
					}

				}

				ans = append(ans, targetFile)

			}

		}
	}

	return ans, nil
}

func (m *MergeBot) createManifest(targetPkgDir string,
	atom *specs.RepoScanAtom) (string, error) {
	// Retrive manifest files of existing ebuilds
	existingAtoms, _ := m.TargetResolver.GetPackageVersions(atom.CatPkg)

	files := atom.Files

	// A single package could be defined multiple times in the
	// YAML specs in order to match multiple versions.
	// In that case the TargetResolver could be without the
	// Manifest files of the elaborated files.
	// To avoid missing DIST files we store in memory the files
	// of the new packages in order to merge the list if it's needed.
	elabFiles, present := m.manifestFiles[atom.CatPkg]
	if present {
		files = append(files, elabFiles...)
	}

	m.manifestFiles[atom.CatPkg] = files

	if len(existingAtoms) > 0 {
		for _, a := range existingAtoms {
			files = append(files, a.Files...)
		}
	}

	manifest := NewManifestFile(files)
	manifestFile := filepath.Join(targetPkgDir, "Manifest")

	return manifestFile, manifest.Write(manifestFile)
}

func (m *MergeBot) copyEbuild(sourcePkgDir, targetPkgDir string,
	atom *specs.RepoScanAtom) (string, error) {
	gpkg, err := atom.ToGentooPackage()
	if err != nil {
		return "", err
	}

	ebuildName := fmt.Sprintf("%s-%s.ebuild", atom.Package, gpkg.GetPVR())
	source := filepath.Join(sourcePkgDir, ebuildName)
	target := filepath.Join(targetPkgDir, ebuildName)

	return target, helpers.CopyFile(source, target)
}
