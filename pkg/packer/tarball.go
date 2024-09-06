/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package packer

import (
	"fmt"
	"os"
	"path/filepath"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	tarf_tools "github.com/geaaru/tar-formers/pkg/tools"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (p *Packer) createTarball(rootfsdir string, out *specs.JobOutput) error {
	var err error
	// Create target directory if doesn't exist
	if !utils.Exists(out.Dir) {
		err = os.MkdirAll(out.Dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	p.Logger.Info(
		":construction:Creating tarball...",
	)

	cfg := tarf_specs.NewConfig(p.Config.Viper)
	// TODO: Permits to customize tarformers
	//       options from s.Config
	cfg.GetLogging().Level = "warning"

	tarformers := executor.NewTarFormers(cfg)
	tarfspec := out.TarformersSpec
	if tarfspec == nil {
		tarfspec = tarf_specs.NewSpecFile()
		// Enable this over a container could generate warnings
		// Ignoring xattr mtime not supported by the underlying filesystem: operation not supported
		tarfspec.SameChtimes = false
	}

	if tarfspec.Writer == nil {
		tarfspec.Writer = tarf_specs.NewWriter()
	}
	// Add rootfs directory
	tarfspec.Writer.AddDir(rootfsdir)

	useExt4compression := true
	opts := tarf_tools.NewTarCompressionOpts(useExt4compression)
	defer opts.Close()

	tarball := filepath.Join(out.Dir, out.Name)
	err = tarf_tools.PrepareTarWriter(tarball, opts)
	if err != nil {
		return err
	}

	if opts.CompressWriter != nil {
		tarformers.SetWriter(opts.CompressWriter)
	} else {
		tarformers.SetWriter(opts.FileWriter)
	}

	err = tarformers.RunTaskWriter(tarfspec)
	if err != nil {
		return err
	}

	p.Logger.Info(
		fmt.Sprintf(
			":factory:File %s created!",
			out.Name))

	return nil
}
