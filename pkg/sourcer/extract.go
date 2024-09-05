/*
Copyright © 2024 Macaroni OS Linux
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
	// TODO: Add filters support on metro specs
	tarfspec := tarf_specs.NewSpecFile()

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

	s.Logger.InfoC(fmt.Sprintf(
		"Archive %s extracted correctly.",
		filepath.Base(source.Target),
	))

	return nil
}
