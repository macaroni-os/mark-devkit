/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"fmt"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type BaseGenerator struct {
	Opts map[string]string
}

func NewBaseGenerator(opts map[string]string) *BaseGenerator {
	return &BaseGenerator{
		Opts: opts,
	}
}

func (g *BaseGenerator) GetOpts() map[string]string {
	return g.Opts
}

func (g *BaseGenerator) setVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref

	delete(values, "versions")

	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)
	renderedArtefacts := []*specs.AutogenArtefact{}

	// Manage static artefacts definition or defined artefacts
	// on process phase
	if len(artefacts) > 0 && (atom.IgnoreArtefacts == nil || !*atom.IgnoreArtefacts) {

		for _, art := range artefacts {
			name, err := helpers.RenderContentWithTemplates(
				art.Name,
				"", "", "asset.name", values, []string{},
			)
			if err != nil {
				return err
			}

			url := ""
			if len(art.SrcUri) > 0 {
				url, err = helpers.RenderContentWithTemplates(
					art.SrcUri[0],
					"", "", "asset.url", values, []string{},
				)
				if err != nil {
					return err
				}
			}

			renderedArtefacts = append(renderedArtefacts, &specs.AutogenArtefact{
				SrcUri: []string{url},
				Use:    art.Use,
				Name:   name,
			})
		}

	}

	// Manage assets defined.
	if atom.HasAssets() {
		for _, asset := range atom.Assets {
			name, err := helpers.RenderContentWithTemplates(
				asset.Name,
				"", "", "asset.name", values, []string{},
			)
			if err != nil {
				return err
			}

			if asset.Prefix == "" && asset.Url == "" {
				return fmt.Errorf(
					"Asset %s for atom %s without prefix and url not admitted",
					asset.Name, atom.Name)
			}

			var srcUri string

			if asset.Prefix != "" {
				srcUri, err = helpers.RenderContentWithTemplates(
					asset.Prefix,
					"", "", "asset.prefix", values, []string{},
				)
				if err != nil {
					return err
				}

				if !strings.HasSuffix(srcUri, "/") {
					srcUri += "/"
				}
				srcUri += name

			} else {
				srcUri, err = helpers.RenderContentWithTemplates(
					asset.Url,
					"", "", "asset.url", values, []string{},
				)
				if err != nil {
					return err
				}

			}

			if log.Config.GetGeneral().Debug {
				log.Debug(fmt.Sprintf("[%s] For asset %s using url %s",
					atom.Name, name, srcUri))
			}

			renderedArtefacts = append(renderedArtefacts, &specs.AutogenArtefact{
				SrcUri: []string{srcUri},
				Use:    asset.Use,
				Name:   name,
			})
		}

	}

	values["artefacts"] = renderedArtefacts

	return nil
}
