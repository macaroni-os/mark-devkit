/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package artefacts

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func DownloadArtefact(
	restGuard *guard.RestGuard,
	atom *specs.AutogenAtom,
	atomUrl, tarballName, downloadDir string) (*specs.RepoScanFile, error) {
	log := logger.GetDefaultLogger()

	ans := &specs.RepoScanFile{
		SrcUri: []string{atomUrl},
		Name:   tarballName,
		Hashes: make(map[string]string, 0),
	}

	uri, err := url.Parse(atomUrl)
	if err != nil {
		return nil, err
	}

	ssl := false

	switch uri.Scheme {
	case "https":
		ssl = true
	default:
		ssl = false
	}

	if uri.Scheme == "ftp" {
		return nil, fmt.Errorf("Not yet implemented!")
	}

	downloadedFilePath := filepath.Join(downloadDir, tarballName)

	node := guard_specs.NewRestNode(uri.Host,
		uri.Host+filepath.Dir(uri.Path), ssl)

	resource := filepath.Base(uri.Path)

	service := guard_specs.NewRestService(uri.Host)
	service.Retries = 3
	service.AddNode(node)

	// Try to use local tarball if available
	if utils.Exists(downloadedFilePath) {
		// Try to retrieve the size of the tarball with HEAD command.
		size, err := RetrieveArtefactSize(restGuard, service, resource)
		if err != nil {
			return nil, err
			log.DebugC(
				fmt.Sprintf(
					"[%s] Error on retrieve artifact size for tarball %s: %s. Ignore tarball.",
					atom.Name, downloadedFilePath, err.Error(), size,
				))
		} else {
			fileReader, err := helpers.GetFileHashes(downloadedFilePath)
			if err != nil {
				return nil, err
			}

			// NOTE: Github tarball API doesn't supply a valid
			//       Content-Length with HEAD method because the
			//       generation of the tarball is in realtime.
			//       This means that will be always re-fetched every time.
			//       We can supply the size through the AutogenArtefact object
			//       in the near future as alternative.
			if fileReader.Size() == size {

				log.DebugC(
					fmt.Sprintf("[%s] Using local tarball %s of size %d.",
						atom.Name, downloadedFilePath, size,
					))

				ans.Hashes["sha512"] = fileReader.Sha512()
				ans.Hashes["blake2b"] = fileReader.Blake2b()
				ans.Size = fmt.Sprintf("%d", fileReader.Size())

				return ans, nil
			} else {

				log.DebugC(
					fmt.Sprintf(
						"[%s] Local tarball %s is with different size (%d != %d). Ignore tarball.",
						atom.Name, downloadedFilePath, fileReader.Size(), size,
					))
			}
		}
	}

	t := service.GetTicket()
	defer t.Rip()

	_, err = restGuard.CreateRequest(t, "GET", "/"+resource)
	if err != nil {
		return nil, err
	}

	artefact, err := restGuard.DoDownload(t, downloadedFilePath)
	if err != nil {
		if t.Response != nil {
			return nil, fmt.Errorf("%s - %s", err.Error(), t.Response.Status)
		} else {
			return nil, fmt.Errorf("%s", err.Error())
		}
	}

	ans.Hashes["sha512"] = artefact.Sha512
	ans.Hashes["blake2b"] = artefact.Blake2b
	ans.Size = fmt.Sprintf("%d", artefact.Size)

	return ans, nil
}

func RetrieveArtefactSize(restGuard *guard.RestGuard,
	service *guard_specs.RestService,
	path string) (int64, error) {

	t := service.GetTicket()
	defer t.Rip()

	_, err := restGuard.CreateRequest(t, "HEAD", "/"+path)
	if err != nil {
		return 0, err
	}

	err = restGuard.Do(t)
	if err != nil {
		return 0, err
	}

	if t.Response.StatusCode == 200 {
		return t.Response.ContentLength, nil
	}

	if t.Response == nil {
		return 0, fmt.Errorf("Invalid response received.")
	}

	return 0, fmt.Errorf("Received response %s", t.Response.Status)
}
