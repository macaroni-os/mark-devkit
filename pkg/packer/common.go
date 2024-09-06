/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package packer

import (
	"fmt"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Packer struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger
}

func NewPacker(c *specs.MarkDevkitConfig) *Packer {
	return &Packer{
		Config: c,
		Logger: log.GetDefaultLogger(),
	}
}

func (p *Packer) Produce(rootfsdir string, out *specs.JobOutput) error {
	var err error
	switch out.Type {
	case "file":
		err = p.createTarball(rootfsdir, out)
	default:
		err = fmt.Errorf("Output type %s not supported!", out.Type)
	}
	return err
}
