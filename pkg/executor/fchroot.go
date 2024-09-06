/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type FchrootOpts struct {
	Verbose     bool
	Debug       bool
	PreserveEnv bool
	Cpu         string
	NoBind      bool
}

type FchrootExecutor struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	Opts       *FchrootOpts
	Quiet      bool
	Entrypoint []string
}

func NewFchrootOpts() *FchrootOpts {
	return &FchrootOpts{
		Verbose:     true,
		Debug:       false,
		PreserveEnv: true,
		Cpu:         "",
		NoBind:      false,
	}
}

func (o *FchrootOpts) GetFlags(binds map[string]string) []string {
	ans := []string{}

	if o.Verbose {
		ans = append(ans, "-v")
	}
	if o.Debug {
		ans = append(ans, "--debug")
	}
	if o.PreserveEnv {
		ans = append(ans, "--preserve-env")
	}
	if o.Cpu != "" {
		ans = append(ans, "--cpu="+o.Cpu)
	}
	if o.NoBind {
		ans = append(ans, "--nobind")
	}

	for k, v := range binds {
		ans = append(ans,
			fmt.Sprintf("--bind=%s:%s", k, v))
	}

	return ans
}

func NewFchrootExecutor(c *specs.MarkDevkitConfig, o *FchrootOpts) *FchrootExecutor {
	return &FchrootExecutor{
		Config: c,
		Logger: log.GetDefaultLogger(),
		Opts:   o,
		Quiet:  false,
	}
}

func (f *FchrootExecutor) RunCommandWithOutput(
	command string, envs map[string]string,
	outBuffer, errBuffer io.WriteCloser,
	entrypoint []string,
	rootfsdir string,
	binds map[string]string) (int, error) {

	ans := 1

	if outBuffer == nil {
		return ans, fmt.Errorf("Invalid outBuffer")
	}
	if errBuffer == nil {
		return ans, fmt.Errorf("Invalid errBuffer")
	}

	fchroot := utils.TryResolveBinaryAbsPath("fchroot")
	fchrootEntrypoint := []string{fchroot}

	// Add fchroot flags
	fchrootEntrypoint = append(fchrootEntrypoint, f.Opts.GetFlags(binds)...)

	// Add rootfs path
	fchrootEntrypoint = append(fchrootEntrypoint, rootfsdir)

	cmds := []string{}
	if len(entrypoint) > 0 {
		cmds = append(cmds, entrypoint...)
	}
	cmds = append(cmds, command)

	cmds = append(fchrootEntrypoint, cmds...)
	chrootCommand := exec.Command(cmds[0], cmds[1:]...)

	if !f.Quiet {
		f.Logger.InfoC(
			f.Logger.Aurora.Bold(
				f.Logger.Aurora.BrightCyan(
					fmt.Sprintf(
						":high-speed_train:>>> fchroot executing...\n- entrypoint: %s\n- command: [%s]",
						fchrootEntrypoint, command))))
	}

	// Convert envs to array list
	elist := os.Environ()
	for k, v := range envs {
		elist = append(elist, k+"="+v)
	}

	chrootCommand.Stdout = outBuffer
	chrootCommand.Stderr = errBuffer
	chrootCommand.Env = elist

	err := chrootCommand.Start()
	if err != nil {
		f.Logger.Error("Error on start command: " + err.Error())
		return 1, err
	}

	err = chrootCommand.Wait()
	if err != nil {
		f.Logger.Error("Error on waiting command: " + err.Error())
		return 1, err
	}

	ans = chrootCommand.ProcessState.ExitCode()

	f.Logger.DebugC(f.Logger.Aurora.Bold(
		f.Logger.Aurora.BrightCyan(
			fmt.Sprintf(":high-speed_train: Exiting [%d]", ans))))

	return ans, nil
}
