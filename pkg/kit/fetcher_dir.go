/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type FetcherDir struct {
	*FetcherCommon
}

func NewFetcherDir(c *specs.MarkDevkitConfig) *FetcherDir {
	return &FetcherDir{
		FetcherCommon: NewFetcherCommon(c),
	}
}

func (f *FetcherDir) Sync(specfile string, opts *FetchOpts) error {
	// Load MergeKit specs
	mkit := specs.NewDistfilesSpec()

	if opts.CleanWorkingDir {
		defer os.RemoveAll(f.GetReposcanDir())
		defer os.RemoveAll(f.GetTargetDir())
	}

	err := mkit.LoadFile(specfile)
	if err != nil {
		return err
	}

	err = f.PrepareSourcesKits(mkit, opts)
	if err != nil {
		return err
	}

	err = f.syncAtoms(mkit, opts)
	if err != nil {
		return err
	}

	return nil
}

func (f *FetcherDir) syncAtoms(mkit *specs.DistfilesSpec, opts *FetchOpts) error {

	// Prepare download directory
	err := helpers.EnsureDirWithoutIds(f.GetDownloadDir(), 0755)
	if err != nil {
		return err
	}

	// Create gentoo packages for filters
	filters := []*gentoo.GentooPackage{}
	if len(opts.Atoms) > 0 {
		for _, atomstr := range opts.Atoms {
			gp, err := gentoo.ParsePackageStr(atomstr)
			if err != nil {
				return fmt.Errorf(
					"invalid atom filter %s: %s",
					atomstr, err.Error())
			}
			filters = append(filters, gp)
		}
	}

	for catpkg, atoms := range f.Resolver.Map {

		f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", catpkg))

		for idx := range atoms {

			if len(filters) > 0 {
				atomGp, err := gentoo.ParsePackageStr(atoms[idx].Atom)
				if err != nil {
					return fmt.Errorf(
						"unexpected error on parse %s: %s",
						atoms[idx].Atom, err.Error())
				}

				admitted := false
				for idx := range filters {
					ok, _ := filters[idx].Admit(atomGp)
					if ok {
						admitted = true
						break
					}
				}

				if !admitted {
					f.Logger.Debug(fmt.Sprintf(
						":factory:[%s] Package filtered. Skipped.",
						atoms[idx].Atom))
					continue
				}
			}

			f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", atoms[idx].Atom))

			f.Stats.IncrementElab()

			if len(atoms[idx].Files) > 0 {
				err := f.syncAtom(mkit, opts, &atoms[idx])
				if err != nil {
					f.AddAtomInError(&atoms[idx], err)
				}
			} else {
				f.Logger.Debug(fmt.Sprintf(":smiling_face_with_sunglasses:[%s] Nothing to do.", atoms[idx].Atom))
			}
		}
	}

	return nil
}

func (f *FetcherDir) syncAtom(mkit *specs.DistfilesSpec, opts *FetchOpts, atom *specs.RepoScanAtom) error {

	toDownload := false
	filesMap := make(map[string]string, 0)
	atomSize := int64(0)

	for _, file := range atom.Files {
		// Check if the file is already present
		downloadedFilePath := filepath.Join(f.GetDownloadDir(), file.Name)
		if _, present := filesMap[file.Name]; !present {
			size, _ := strconv.ParseInt(file.Size, 10, 64)
			atomSize += size
		}

		filesMap[file.Name] = file.Size

		if !utils.Exists(downloadedFilePath) {
			toDownload = true
			// Skip break to get all files size
			// break
		}
	}

	if toDownload {
		f.Logger.InfoC(
			fmt.Sprintf(":factory:[%s] Downloading files...", atom.Atom))

		err := f.DownloadAtomsFiles(mkit, atom)
		if err != nil {
			f.Stats.IncrementErrors()
			return err
		}

		f.Stats.IncrementAtoms()

		f.Logger.InfoC(
			fmt.Sprintf(":medal: [%s] Files synced.", atom.Atom))
	} else {
		f.Logger.InfoC(
			fmt.Sprintf(":medal: [%s] Already synced.", atom.Atom))
	}

	if atomSize > 0 {
		f.Stats.IncrementSize(atomSize)
	}

	return nil
}
