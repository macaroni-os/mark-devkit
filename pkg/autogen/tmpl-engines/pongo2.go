/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package tmplengine

import (
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/flosch/pongo2/v6"
)

type Pongo2TemplateEngine struct {
	*CoreTemplateEngine
}

func NewPongo2TemplateEngine() *Pongo2TemplateEngine {
	return &Pongo2TemplateEngine{
		CoreTemplateEngine: &CoreTemplateEngine{},
	}
}

func (t *Pongo2TemplateEngine) Render(aspec *specs.AutogenSpec,
	atom, def *specs.AutogenAtom, valref *map[string]interface{},
	targetFile string) error {

	values := *valref
	templateFilePath := t.GetTemplateFile(aspec, atom, def)

	// Read the template file
	data, err := os.ReadFile(templateFilePath)
	if err != nil {
		return err
	}

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromBytes(data)
	if err != nil {
		return err
	}

	// Render the content
	content, err := tpl.ExecuteBytes(pongo2.Context(values))
	if err != nil {
		return err
	}

	// Write the file
	return os.WriteFile(targetFile, []byte(content), 0644)
}
