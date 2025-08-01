/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"gopkg.in/yaml.v3"
)

type CustomGenerator struct {
	Opts map[string]string
}

type CustomGeneratorValues struct {
	Name      string                  `json:"name" yaml:"name"`
	Atom      *specs.AutogenAtom      `json:"atom,omitempty" yaml:"atom,omitempty"`
	Values    *map[string]interface{} `json:"vars,omitempty" yaml:"vars,omitempty"`
	Artefacts []*specs.AutogenAsset   `json:"artefacts,omitempty" yaml:"artefacts,omitempty"`
}

func (v *CustomGeneratorValues) WriteYamlFile(f string) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func ReadVars(file string) (*CustomGeneratorValues, error) {
	ans := &CustomGeneratorValues{}

	content, err := os.ReadFile(file)
	if err != nil {
		return ans, err
	}

	err = yaml.Unmarshal(content, ans)
	return ans, err
}

func NewCustomGenerator(opts map[string]string) *CustomGenerator {
	return &CustomGenerator{
		Opts: opts,
	}
}

func (g *CustomGenerator) GetType() string {
	return specs.GeneratorCustom
}

func (g *CustomGenerator) GetOpts() map[string]string {
	return g.Opts
}

func (g *CustomGenerator) GetElabPaths(pkgname string) (string, string, string, string) {

	// Create the temporary directory of the package
	workdir, _ := g.Opts["workdir"]
	specfile, _ := g.Opts["specfile"]
	script, _ := g.Opts["script"]

	pkgWorkDir := filepath.Join(workdir, "custom-generator", pkgname)
	pkgVarfile := filepath.Join(pkgWorkDir, "input.yml")
	pkgVersionsfile := filepath.Join(pkgWorkDir, "versions.yml")

	if !filepath.IsAbs(script) {
		script = filepath.Join(filepath.Dir(specfile), script)
	}

	return pkgWorkDir, pkgVarfile, pkgVersionsfile, script
}

func (g *CustomGenerator) runScript(cmds []string,
	atom *specs.AutogenAtom,
	pkgVarfile, pkgVersionsfile string,
	mapref *map[string]interface{}) (*map[string]interface{}, error) {

	log := logger.GetDefaultLogger()
	values := *mapref

	input := &CustomGeneratorValues{
		Name:   atom.Name,
		Atom:   atom,
		Values: mapref,
	}

	// Write input file
	err := input.WriteYamlFile(pkgVarfile)
	if err != nil {
		return nil, err
	}

	command := exec.Command(cmds[0], cmds[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		log.Error("Error on start command: " + err.Error())
		return nil, err
	}

	err = command.Wait()
	if err != nil {
		log.Error("Error on waiting command: " + err.Error())
		return nil, err
	}

	exitCode := command.ProcessState.ExitCode()

	if exitCode != 0 {
		return nil, fmt.Errorf("Script %s exiting with %d. Stop elaboration.",
			cmds[0], exitCode)
	}

	// Read versions file
	output, err := ReadVars(pkgVersionsfile)
	if err != nil {
		return nil, err
	}

	if output.Values != nil {
		err = helpers.SanitizeMapVersionsField(atom.Name, output.Values)
		if err != nil {
			return nil, err
		}
	}

	// Merge output values in existing values and on atomMap
	for k, v := range *output.Values {
		values[k] = v
	}

	if len(output.Artefacts) > 0 {
		artefacts := []*specs.AutogenArtefact{}

		for _, asset := range output.Artefacts {
			artefacts = append(artefacts, &specs.AutogenArtefact{
				SrcUri: []string{asset.Url},
				Use:    asset.Use,
				Name:   asset.Name,
			})
		}

		values["artefacts"] = artefacts
	}

	return &values, nil
}

func (g *CustomGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {
	values := *mapref

	enableSetVersion, available := g.Opts["enable_set_version"]
	if !available {
		enableSetVersion = "true"
	}

	delete(values, "versions")

	if enableSetVersion == "true" {
		_, pkgVarfile, pkgVersionsfile, script := g.GetElabPaths(atom.Name)

		// Execute defined script that have three arguments
		// ./custom-script.sh <mode> <var-files> <output-vars-files>
		// <mode> := process | set-version
		cmds := []string{
			script,
			"set-version",
			pkgVarfile,
			pkgVersionsfile,
		}

		ans, err := g.runScript(cmds, atom, pkgVarfile, pkgVersionsfile, mapref)
		if err != nil {
			return err
		}

		for k, v := range *ans {
			values[k] = v
		}
	}

	// If there artefacts apply render to asset name and url using the values.
	// Always execute these operations also when enable_set_version is set to false.

	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)
	renderedArtefacts := []*specs.AutogenArtefact{}
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

	// Add atom assets as artefacts
	if atom.HasAssets() {
		for _, asset := range atom.Assets {
			name, err := helpers.RenderContentWithTemplates(
				asset.Name,
				"", "", "asset.name", values, []string{},
			)
			if err != nil {
				return err
			}

			url := ""
			if asset.Url != "" {
				// POST: We use the url value as urlBase
				url, err = helpers.RenderContentWithTemplates(
					asset.Url,
					"", "", "asset.url", values, []string{},
				)
				if err != nil {
					return err
				}

			}

			renderedArtefacts = append(renderedArtefacts, &specs.AutogenArtefact{
				SrcUri: []string{url},
				Use:    asset.Use,
				Name:   name,
			})
		}
	}

	values["artefacts"] = renderedArtefacts

	return nil
}

// Retrieve metadata and all availables tags/releases
func (g *CustomGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	_, available := g.Opts["script"]
	if !available {
		return nil, fmt.Errorf("For atom %s the generator is without script option!",
			atom.Name)
	}

	pkgWorkDir, pkgVarfile, pkgVersionsfile, script := g.GetElabPaths(atom.Name)

	err := os.MkdirAll(pkgWorkDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Execute defined script that have three arguments
	// ./custom-script.sh <mode> <var-files> <output-vars-files>
	// <mode> := process | set-version
	cmds := []string{
		script,
		"process",
		pkgVarfile,
		pkgVersionsfile,
	}

	m := make(map[string]interface{}, 0)
	ans, err := g.runScript(cmds, atom, pkgVarfile, pkgVersionsfile, &m)
	if err != nil {
		return nil, err
	}

	return ans, nil
}
