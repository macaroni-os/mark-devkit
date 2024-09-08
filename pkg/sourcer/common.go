/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package sourcer

import (
	"fmt"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

type SourceHandler struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger
}

func NewSourceHandler(c *specs.MarkDevkitConfig) *SourceHandler {
	return &SourceHandler{
		Config: c,
		Logger: log.GetDefaultLogger(),
	}
}

func (s *SourceHandler) Produce(source *specs.JobSource) error {
	var err error
	switch source.Type {
	case "http":
		err = s.fetchttp(source)
	case "anise":
		s.Logger.Debug("Nothing to do on produce phase for anise source.")
	default:
		err = fmt.Errorf("Source type %s not supported!", source.Type)
	}
	return err
}

func (s *SourceHandler) Consume(source *specs.JobSource, rootfsdir string) error {
	var err error
	switch source.Type {
	case "http":
		err = s.extract(source, rootfsdir)
	case "anise":
		err = s.aniseConsume(source, rootfsdir)
	default:
		err = fmt.Errorf("Source type %s not supported!", source.Type)
	}

	return err
}
