/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"path/filepath"
)

func (a *AutogenBot) syncTarballs(opts *AutogenBotOpts) error {

	if a.Fetcher.GetType() == "dir" {
		// Nothing to do
		return nil
	}

	// Retrieve the files already on server
	files, err := a.Fetcher.GetFilesList()
	if err != nil {
		return err
	}

	// Create a map of the files
	serverfilesMap := make(map[string]bool, 0)
	filesMap := make(map[string]bool, 0)
	for idx := range files {
		serverfilesMap[files[idx]] = true
	}
	downloadDir := a.GetDownloadDir()

	for idx := range a.ElabAtoms {

		atom := a.ElabAtoms[idx]

		for _, file := range atom.Files {

			targetFilePath := a.Fetcher.GetFilePath(file.Name)
			// Check if the files is already present on server
			if _, present := serverfilesMap[targetFilePath]; present {
				a.Logger.Debug(fmt.Sprintf(
					":brain:[%s] Tarball %s already present on server.",
					atom.Atom, file.Name))
				continue
			}

			// Check if the file is already present
			if _, present := filesMap[file.Name]; present {
				// File already processed
				continue
			}
			filesMap[file.Name] = true

			downloadedFilePath := filepath.Join(downloadDir, file.Name)
			err := a.Fetcher.SyncFile(atom.Atom, downloadedFilePath, file.Name, &file.Hashes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
