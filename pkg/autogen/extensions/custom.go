/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/autogen/generators"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
)

type ExtensionCustom struct {
	Opts map[string]string
}

func NewExtensionCustom(opts map[string]string) (*ExtensionCustom, error) {
	_, present := opts["script"]
	if !present {
		return nil, fmt.Errorf("script option not defined")
	}

	return &ExtensionCustom{
		Opts: opts,
	}, nil
}

func (e *ExtensionCustom) GetName() string { return specs.ExtensionCustom }
func (e *ExtensionCustom) GetOpts() map[string]string {
	return e.Opts
}

func (e *ExtensionCustom) GetElabPaths(pkgname string) (string, string, string, string) {

	// Create the temporary directory of the package
	workdir, _ := e.Opts["workdir"]
	specfile, _ := e.Opts["specfile"]
	script, _ := e.Opts["script"]

	pkgWorkDir := filepath.Join(workdir, "custom-extension", pkgname)
	pkgVarfile := filepath.Join(pkgWorkDir, "input.yml")
	pkgOutputfile := filepath.Join(pkgWorkDir, "output.yml")

	if !filepath.IsAbs(script) {
		script = filepath.Join(filepath.Dir(specfile), script)
	}

	pkgWorkDir, _ = filepath.Abs(pkgWorkDir)
	pkgVarfile, _ = filepath.Abs(pkgVarfile)
	pkgOutputfile, _ = filepath.Abs(pkgOutputfile)

	return pkgWorkDir, pkgVarfile, pkgOutputfile, script
}

func (e *ExtensionCustom) Elaborate(restGuard *guard.RestGuard,
	atom, def *specs.AutogenAtom,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref
	pkgWorkDir, pkgVarfile, pkgOutputfile, script := e.GetElabPaths(atom.Name)

	err := os.MkdirAll(pkgWorkDir, os.ModePerm)
	if err != nil {
		return err
	}

	input := &generators.CustomGeneratorValues{
		Name:   atom.Name,
		Atom:   atom,
		Values: mapref,
	}

	downloadDir := e.Opts["download_dir"]
	workdir := e.Opts["workdir"]
	if !filepath.IsAbs(downloadDir) {
		downloadDir, _ = filepath.Abs(downloadDir)
	}
	if !filepath.IsAbs(workdir) {
		workdir, _ = filepath.Abs(workdir)
	}
	values["workdir"] = workdir
	values["download_dir"] = downloadDir
	values["specfile"] = e.Opts["specfile"]
	values["mirror"] = e.Opts["mirror"]

	if values["mirror"] == "" {
		values["mirror"] = "mirror://macaroni"
	}

	// Ensure download dir. Could be not present the first time.
	err = os.MkdirAll(e.Opts["download_dir"], os.ModePerm)
	if err != nil {
		return err
	}

	// Write input file
	err = input.WriteYamlFile(pkgVarfile)
	if err != nil {
		return err
	}

	// Execute defined script that have three arguments
	// ./custom-script.sh <var-files> <output-vars-files>
	cmds := []string{
		script,
		pkgVarfile,
		pkgOutputfile,
	}

	command := exec.Command(cmds[0], cmds[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		log.Error("Error on start command: " + err.Error())
		return err
	}

	err = command.Wait()
	if err != nil {
		log.Error("Error on waiting command: " + err.Error())
		return err
	}

	exitCode := command.ProcessState.ExitCode()

	if exitCode != 0 {
		return fmt.Errorf("Script %s exiting with %d. Stop elaboration.",
			cmds[0], exitCode)
	}

	// Read output file
	output, err := generators.ReadVars(pkgOutputfile)
	if err != nil {
		return err
	}

	// Merge output values in existing values and on atomMap
	if output.Values != nil {
		for k, v := range *output.Values {
			values[k] = v
		}
	}

	if len(output.Artefacts) > 0 {
		artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)

		for _, asset := range output.Artefacts {
			artefacts = append(artefacts, &specs.AutogenArtefact{
				SrcUri: []string{asset.Url},
				Use:    asset.Use,
				Name:   asset.Name,
				Local:  asset.Local,
			})
		}

		values["artefacts"] = artefacts
	}

	delete(values, "workdir")
	delete(values, "download_dir")
	delete(values, "specfile")
	delete(values, "mirror")

	if !logger.GetDefaultLogger().Config.GetGeneral().Debug {
		defer os.RemoveAll(filepath.Join(workdir, "custom-extension"))
	}

	return nil
}
