/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Fetcher interface {
	Sync(specfile string, opts *FetchOpts) error
	SyncFile(name, source, target string, hashes *map[string]string) error
	SetWorkDir(d string)
	SetDownloadDir(d string)
	GetWorkDir() string
	GetTargetDir() string
	GetReposcanDir() string
	GetDownloadDir() string
	GetFilePath(f string) string
	GetFilesList() ([]string, error)
	GetType() string
	GetStats() *AtomsStats
	GetAtomsInError() *[]*AtomError
}

type FetchOpts struct {
	Concurrency     int
	GenReposcan     bool
	Verbose         bool
	CleanWorkingDir bool
	CheckOnlySize   bool

	Atoms []string
}

func NewFetchOpts() *FetchOpts {
	return &FetchOpts{
		Concurrency:     10,
		GenReposcan:     true,
		Verbose:         false,
		CleanWorkingDir: false,
		CheckOnlySize:   false,
		Atoms:           []string{},
	}
}

func NewFetcher(c *specs.MarkDevkitConfig, backend string, opts map[string]string) (Fetcher, error) {
	switch backend {
	case "dir":
		return NewFetcherDir(c), nil
	case "s3":
		return NewFetcherS3(c, opts)
	default:
		return nil, fmt.Errorf("invalid fetcher backend %s", backend)
	}
}
