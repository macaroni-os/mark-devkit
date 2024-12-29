/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Fetcher interface {
	Sync(specfile string, opts *FetchOpts) error
	SetWorkDir(d string)
	GetWorkDir() string
	GetTargetDir() string
	GetReposcanDir() string
	GetDownloadDir() string
	GetStats() *AtomsStats
	GetAtomsInError() *[]*AtomError
}

type FetchOpts struct {
	Concurrency     int
	GenReposcan     bool
	Verbose         bool
	CleanWorkingDir bool
}

func NewFetchOpts() *FetchOpts {
	return &FetchOpts{
		Concurrency:     10,
		GenReposcan:     true,
		Verbose:         false,
		CleanWorkingDir: false,
	}
}

func NewFetcher(c *specs.MarkDevkitConfig, backend string) (Fetcher, error) {
	switch backend {
	case "dir":
		return NewFetcherDir(c), nil
	case "s3":
		return nil, fmt.Errorf("backend %s not yet implemented", backend)
	default:
		return nil, fmt.Errorf("invalid fetcher backend %s", backend)
	}
}