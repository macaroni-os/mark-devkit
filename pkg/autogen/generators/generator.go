/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package generators

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Generator interface {
	Process(atom *specs.AutogenAtom) (*map[string]interface{}, error)
	GetType() string
	SetVersion(atom *specs.AutogenAtom, v string, mapref *map[string]interface{}) error
}

func NewGenerator(t string, opts map[string]string) (Generator, error) {
	switch t {
	case specs.GeneratorBuiltinGitub:
		return NewGithubGenerator(), nil
	case specs.GeneratorBuiltinDirListing:
		return NewDirlistingGenerator(opts), nil
	case specs.GeneratorBuiltinNoop:
		return NewNoopGenerator(), nil
	case specs.GeneratorBuiltinPypi:
		return NewPypiGenerator(), nil
	case specs.GeneratorCustom:
		return NewCustomGenerator(opts), nil
	default:
		return nil, fmt.Errorf("Invalid generator type %s", t)
	}
}
