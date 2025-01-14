/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package tmplengine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/macaroni-os/macaronictl/pkg/utils"
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"gopkg.in/yaml.v3"
)

type J2cliTemplateEngine struct {
	*CoreTemplateEngine
	Opts []string
}

func NewJ2CliTemplateEngine(opts []string) *J2cliTemplateEngine {
	return &J2cliTemplateEngine{
		CoreTemplateEngine: &CoreTemplateEngine{},
		Opts:               opts,
	}
}

func (t *J2cliTemplateEngine) Render(aspec *specs.AutogenSpec,
	atom, def *specs.AutogenAtom, valref *map[string]interface{},
	targetFile string) error {

	templateFilePath := t.GetTemplateFile(aspec, atom, def)
	logger := log.GetDefaultLogger()

	// Create temporary directory where store vars files
	tmpdir, err := os.MkdirTemp("", fmt.Sprintf("mark-devkit-%s", atom.Name))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	// Create var files
	varsFile := filepath.Join(tmpdir, "data.yml")
	d, err := yaml.Marshal(valref)
	if err != nil {
		return err
	}

	err = os.WriteFile(varsFile, d, 0644)
	if err != nil {
		return err
	}

	// Command to execute:
	// j2 template.j2 data.yml -o destfile
	args := []string{
		templateFilePath, varsFile,
		"-o", targetFile,
	}

	if len(t.Opts) > 0 {
		args = append(args, t.Opts...)
	}
	j2 := utils.TryResolveBinaryAbsPath("j2")
	j2Command := exec.Command(j2, args...)

	if logger.Config.GetGeneral().Debug {
		j2Command.Stdout = os.Stdout
		j2Command.Stderr = os.Stderr
	}

	// Write the file
	return j2Command.Run()
}
