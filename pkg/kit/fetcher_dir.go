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

	for catpkg, atoms := range f.Resolver.Map {

		f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", catpkg))

		for idx := range atoms {
			f.Logger.Debug(fmt.Sprintf(":factory:[%s] Analyzing ...", atoms[idx].Atom))

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

	// NOTE: At the moment we use old flat-dir mode only.

	for _, file := range atom.Files {
		// Check if the file is already present
		downloadedFilePath := filepath.Join(f.GetDownloadDir(), file.Name)
		if !utils.Exists(downloadedFilePath) {
			toDownload = true
			break
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

	f.Stats.IncrementElab()

	return nil
}
