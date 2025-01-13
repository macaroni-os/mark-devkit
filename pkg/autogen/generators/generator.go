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
	Process(atom *specs.AutogenAtom, def *specs.AutogenAtom) (*map[string]interface{}, error)
	GetType() string
	SetVersion(atom *specs.AutogenAtom, v string, mapref *map[string]interface{}) error
}

func NewGenerator(t string) (Generator, error) {
	switch t {
	case specs.GeneratorBuiltinGitub:
		return NewGithubGenerator(), nil
	default:
		return nil, fmt.Errorf("Invalid generator type %s", t)
	}
}
