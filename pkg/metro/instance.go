/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package metro

import (
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Metro struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	Envs []*specs.EnvVar
}

func NewMetro(c *specs.MarkDevkitConfig) *Metro {
	return &Metro{
		Config: c,
		Logger: log.GetDefaultLogger(),
		Envs:   []*specs.EnvVar{},
	}
}

func (m *Metro) GetConfig() *specs.MarkDevkitConfig { return m.Config }
func (m *Metro) GetEnvs() *[]*specs.EnvVar          { return &m.Envs }
