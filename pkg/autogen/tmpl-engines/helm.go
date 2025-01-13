/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package tmplengine

import (
	"os"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type HelmTemplateEngine struct {
	*CoreTemplateEngine
}

func NewHelmTemplateEngine() *HelmTemplateEngine {
	return &HelmTemplateEngine{
		CoreTemplateEngine: &CoreTemplateEngine{},
	}
}

func (t *HelmTemplateEngine) Render(aspec *specs.AutogenSpec,
	atom, def *specs.AutogenAtom, valref *map[string]interface{},
	targetFile string) error {

	values := *valref
	templateFilePath := t.GetTemplateFile(aspec, atom, def)

	// Read the template file
	data, err := os.ReadFile(templateFilePath)
	if err != nil {
		return err
	}

	// Render the content
	content, err := helpers.RenderContentWithTemplates(
		string(data),
		"", "", targetFile, values, []string{},
	)
	if err != nil {
		return err
	}

	// Write the file
	return os.WriteFile(targetFile, []byte(content), 0644)
}
