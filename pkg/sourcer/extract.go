/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package sourcer

import (
	"fmt"
	"path/filepath"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	tarf_tools "github.com/geaaru/tar-formers/pkg/tools"
)

func (s *SourceHandler) extract(source *specs.JobSource, rootfsdir string) error {
	cfg := tarf_specs.NewConfig(s.Config.Viper)
	// TODO: Permits to customize tarformers
	//       options from s.Config
	cfg.GetLogging().Level = "warning"

	tarformers := executor.NewTarFormers(cfg)
	// If the specfile is defined i use the
	// value from YAML.
	tarfspec := source.TarformersSpec
	if tarfspec == nil {
		tarfspec = tarf_specs.NewSpecFile()
		// Enable this over a container could generate warnings
		// Ignoring xattr mtime not supported by the underlying filesystem: operation not supported
		tarfspec.SameChtimes = false

		// Set the options from config.
		tarfspec.EnableMutex = s.Config.GetTarFlows().Mutex4Dirs
		tarfspec.MaxOpenFiles = s.Config.GetTarFlows().MaxOpenFiles
		tarfspec.BufferSize = s.Config.GetTarFlows().CopyBufferSize
		tarfspec.Validate = s.Config.GetTarFlows().Validate
	} // else using options from JobSource

	useExt4compression := true
	opts := tarf_tools.NewTarReaderCompressionOpts(useExt4compression)

	err := tarf_tools.PrepareTarReader(source.Target, opts)
	if err != nil {
		return fmt.Errorf("error on prepare reader:", err.Error())
	}

	if opts.CompressReader != nil {
		tarformers.SetReader(opts.CompressReader)
	} else {
		tarformers.SetReader(opts.FileReader)
	}

	err = tarformers.RunTask(tarfspec, rootfsdir)
	opts.Close()
	if err != nil {
		return fmt.Errorf("error on process tarball :" + err.Error())
	}

	s.Logger.Info(fmt.Sprintf(
		"Archive %s extracted correctly.",
		filepath.Base(source.Target),
	))

	return nil
}
