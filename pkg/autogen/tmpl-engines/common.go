/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package tmplengine

import (
	"path/filepath"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type CoreTemplateEngine struct {
	Logger *log.MarkDevkitLogger
}

func (c *CoreTemplateEngine) SetLogger(l *log.MarkDevkitLogger) { c.Logger = l }

func (c *CoreTemplateEngine) GetTemplateFile(aspec *specs.AutogenSpec,
	atom, def *specs.AutogenAtom) string {

	templateFilePath := filepath.Join(filepath.Dir(aspec.File),
		atom.GetTemplate(def))

	return templateFilePath
}
