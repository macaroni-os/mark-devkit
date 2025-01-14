/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package tmplengine

import (
	"fmt"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type TemplateEngine interface {
	Render(aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom, values *map[string]interface{}, targetFile string) error
	SetLogger(l *log.MarkDevkitLogger)
}

func NewTemplateEngine(t string) (TemplateEngine, error) {
	switch t {
	case specs.TmplEngineHelm:
		return NewHelmTemplateEngine(), nil
	case specs.TmplEnginePongo2:
		return NewPongo2TemplateEngine(), nil
	default:
		return nil, fmt.Errorf("Invalid template engine %s", t)
	}
}
