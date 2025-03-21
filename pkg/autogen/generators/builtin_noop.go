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

type NoopGenerator struct{}

func NewNoopGenerator() *NoopGenerator {
	return &NoopGenerator{}
}

func (g *NoopGenerator) GetType() string {
	return specs.GeneratorBuiltinNoop
}

func (g *NoopGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref

	delete(values, "versions")

	artefacts := []*specs.AutogenArtefact{}

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
					"Asset %s for atom %s without prefix and url not admitted with noop",
					asset.Name, atom.Name)
			}

			var srcUri string

			if asset.Prefix != "" {
				srcUri, err = helpers.RenderContentWithTemplates(
					asset.Prefix,
					"", "", "asset.name", values, []string{},
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
					"", "", "asset.name", values, []string{},
				)
				if err != nil {
					return err
				}

			}

			if log.Config.GetGeneral().Debug {
				log.Debug(fmt.Sprintf("[%s] For asset %s using url %s",
					atom.Name, name, srcUri))
			}

			artefacts = append(artefacts, &specs.AutogenArtefact{
				SrcUri: []string{srcUri},
				Use:    asset.Use,
				Name:   name,
			})
		}

	}

	values["artefacts"] = artefacts

	return nil
}

func (g *NoopGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)

	return &ans, nil
}
