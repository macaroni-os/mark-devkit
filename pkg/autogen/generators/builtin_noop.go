/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type NoopGenerator struct {
	*BaseGenerator
}

func NewNoopGenerator() *NoopGenerator {
	opts := make(map[string]string, 0)
	return &NoopGenerator{
		BaseGenerator: NewBaseGenerator(opts),
	}
}

func (g *NoopGenerator) GetType() string {
	return specs.GeneratorBuiltinNoop
}

func (g *NoopGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	return g.BaseGenerator.setVersion(atom, version, mapref)
}

func (g *NoopGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)

	return &ans, nil
}
