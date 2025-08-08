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
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type ExtensionBase struct {
	Opts map[string]string
}

func (e *ExtensionBase) GetOpts() map[string]string {
	return e.Opts
}

func (e *ExtensionBase) cleanup(mapref *map[string]interface{}) {
	values := *mapref

	delete(values, "workdir")
	delete(values, "download_dir")
	delete(values, "mirror")
}

func (e *ExtensionBase) getTarformersConfig() *tarf_specs.Config {
	// Check instance
	config := tarf_specs.NewConfig(nil)
	if logger.GetDefaultLogger().Config.GetGeneral().Debug {
		config.GetLogging().Level = "info"
	} else {
		config.GetGeneral().Debug = false
		config.GetLogging().Level = "info"
	}

	return config
}

func (e *ExtensionBase) unpackArtefact(downloadDir, targetDir string,
	art *specs.RepoScanFile,
	atom, def *specs.AutogenAtom,
	mapref *map[string]interface{}) error {

	tarball := filepath.Join(downloadDir, art.Name)

	// Check instance
	config := tarf_specs.NewConfig(nil)
	if logger.GetDefaultLogger().Config.GetGeneral().Debug {
		config.GetLogging().Level = "info"
	}

	tarformers := executor.NewTarFormers(config)
	s := tarf_specs.NewSpecFile()
	// We don't need to keep the original permission of the files
	// and owner.
	s.SameOwner = false
	s.SameChtimes = false

	tarfOpts := tools.NewTarReaderCompressionOpts(true)
	defer tarfOpts.Close()

	err := tools.PrepareTarReader(tarball, tarfOpts)
	if err != nil {
		return fmt.Errorf("Error on prepare reader:", err.Error())
	}

	if tarfOpts.CompressReader != nil {
		tarformers.SetReader(tarfOpts.CompressReader)
	} else {
		tarformers.SetReader(tarfOpts.FileReader)
	}

	err = tarformers.RunTask(s, targetDir)
	if err != nil {
		return fmt.Errorf("Error on process tarball :" + err.Error())
	}

	// Apply patches to sources if availables
	values := *mapref

	patches, _ := values["patches"].([]interface{})

	if len(patches) > 0 {
		// Retrieve the path of the autogen specs
		// in order to generate the patch path
		filesDir, _ := e.Opts["files_dir"]

		// Retrieve the patch of the sources
		pkgSourceDir := ""
		unpackDirPrefix, _ := e.Opts["unpack_srcdir_prefix"]
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			if unpackDirPrefix != "" {
				if strings.HasPrefix(entry.Name(), unpackDirPrefix) {
					pkgSourceDir = filepath.Join(targetDir, entry.Name())
					break
				}
			} else {
				// POST: take the first directory
				pkgSourceDir = filepath.Join(targetDir, entry.Name())
			}
		}

		for idx := range patches {

			patch, _ := patches[idx].(string)
			patchPath, _ := filepath.Abs(filepath.Join(filesDir, patch))

			patchDone := false
			for _, pflag := range []string{"1", "0"} {
				err := e.doPatch(patchPath, pkgSourceDir, pflag, true)
				if err == nil {
					err = e.doPatch(patchPath, pkgSourceDir, pflag, false)
					if err != nil {
						return fmt.Errorf(
							"error on apply patch %s: %s",
							patch, err.Error())
					}
					patchDone = true
					break
				}
			}

			if !patchDone {
				return fmt.Errorf("patch %s is not usable.", patch)
			}

		}
	}

	return nil
}

func (e *ExtensionBase) doPatch(patch, unpackDir, pflag string, dryRun bool) error {
	log := logger.GetDefaultLogger()
	patchBin := utils.TryResolveBinaryAbsPath("patch")
	args := []string{
		"-p" + pflag,
		"-s",
		//"--verbose",
		//"-d",
		//unpackDir,
		"-i",
		patch,
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	log.DebugC(fmt.Sprintf("Running command patch %s", strings.Join(args, " ")))

	patchCommand := exec.Command(patchBin, args...)
	patchCommand.Dir = unpackDir
	if log.Config.GetGeneral().Debug {
		patchCommand.Stdout = os.Stdout
		patchCommand.Stderr = os.Stderr
	}

	err := patchCommand.Start()
	if err != nil {
		return fmt.Errorf("Error on start patch command: %s", err.Error())
	}

	err = patchCommand.Wait()
	if err != nil {
		return fmt.Errorf("Error on waiting patch command: %s", err.Error())
	}

	if patchCommand.ProcessState.ExitCode() != 0 {
		return fmt.Errorf(
			"patch command exiting with %d",
			patchCommand.ProcessState.ExitCode())
	}

	return nil
}
