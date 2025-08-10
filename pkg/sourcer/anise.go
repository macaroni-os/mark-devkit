/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package sourcer

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/macaroni-os/mark-devkit/pkg/executor"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (s *SourceHandler) aniseConsume(source *specs.JobSource, rootfsdir string) error {
	// Prepare Stdout, Stderr writer
	stdOutWriter := executor.NewExecutorWriter("stdout", false)
	stdErrWriter := executor.NewExecutorWriter("stderr", false)

	envsMaps := make(map[string]string, 0)

	// Create the database inside the chroot and
	// run repo update.
	err := s.aniseCommand([]string{
		"repo", "update",
	}, envsMaps, stdOutWriter, stdErrWriter, rootfsdir,
		source.AniseConfigPath)
	if err != nil {
		return err
	}

	// Install repositories
	if len(source.AniseRepositories) > 0 {
		err = s.aniseCommand(append(
			[]string{
				"install", "-y", "--skip-config-protect",
			}, source.AniseRepositories...),
			envsMaps, stdOutWriter, stdErrWriter, rootfsdir,
			source.AniseConfigPath)

		if err != nil {
			return err
		}
	}

	// Install packages
	if len(source.AnisePackages) > 0 {
		err = s.aniseCommand(append(
			[]string{
				"install", "--sync-repos", "-y", "--skip-config-protect",
			}, source.AnisePackages...),
			envsMaps, stdOutWriter, stdErrWriter, rootfsdir,
			source.AniseConfigPath)

		if err != nil {
			return err
		}
	}

	err = s.aniseCommand([]string{
		"cleanup",
	}, envsMaps, stdOutWriter, stdErrWriter, rootfsdir,
		source.AniseConfigPath)
	if err != nil {
		return err
	}

	return nil
}

func (s *SourceHandler) aniseCommand(args []string,
	envs map[string]string, outBuffer, errBuffer io.WriteCloser,
	rootfsdir, config string) error {

	if outBuffer == nil {
		return fmt.Errorf("Invalid outBuffer")
	}
	if errBuffer == nil {
		return fmt.Errorf("Invalid errBuffer")
	}

	anise := utils.TryResolveBinaryAbsPath("anise")

	elist := os.Environ()
	for k, v := range envs {
		elist = append(elist, k+"="+v)
	}

	aniseOpts := []string{
		"--system-target", rootfsdir,
	}
	if config != "" {
		aniseOpts = append(aniseOpts, []string{
			"--config", config,
		}...)
	}

	args = append(aniseOpts, args...)

	cmd := exec.Command(anise, args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	cmd.Stdin = os.Stdin
	cmd.Env = elist

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error on start command %s: %s",
			anise, err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error on wait command %s: %s",
			anise, err.Error())
	}

	res := cmd.ProcessState.ExitCode()

	if res != 0 {
		return fmt.Errorf("%s exiting with %d!",
			anise, res)
	}

	return nil
}
