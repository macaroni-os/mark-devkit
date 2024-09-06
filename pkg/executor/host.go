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
)

type HostExecutor struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	Entrypoint []string
	Quiet      bool
}

func NewHostExecutor(c *specs.MarkDevkitConfig) *HostExecutor {
	return &HostExecutor{
		Config:     c,
		Logger:     log.GetDefaultLogger(),
		Entrypoint: []string{},
		Quiet:      false,
	}
}

func (h *HostExecutor) RunCommandWithOutput(
	command string, envs map[string]string,
	outBuffer, errBuffer io.WriteCloser,
	entryPoint []string) (int, error) {

	ans := 1

	entrypoint := []string{"/bin/bash", "-c"}
	if len(h.Entrypoint) > 0 {
		entrypoint = h.Entrypoint
	}
	if len(entryPoint) > 0 {
		entrypoint = entryPoint
	}

	if outBuffer == nil {
		return 1, fmt.Errorf("Invalid outBuffer")
	}
	if errBuffer == nil {
		return 1, fmt.Errorf("Invalid errBuffer")
	}

	cmds := append(entrypoint, command)

	hostCommand := exec.Command(cmds[0], cmds[1:]...)

	if !h.Quiet {
		h.Logger.InfoC(
			h.Logger.Aurora.Bold(
				h.Logger.Aurora.BrightYellow(
					fmt.Sprintf(":locomotive:>>> Executing...\n- entrypoint: %s\n- command: [%s]",
						entrypoint, command))))
	}

	// Convert envs to array list
	elist := os.Environ()
	for k, v := range envs {
		elist = append(elist, k+"="+v)
	}

	hostCommand.Stdout = outBuffer
	hostCommand.Stderr = errBuffer
	hostCommand.Stdin = os.Stdin
	hostCommand.Env = elist

	err := hostCommand.Start()
	if err != nil {
		h.Logger.Error("Error on start command: " + err.Error())
		return 1, err
	}

	err = hostCommand.Wait()
	if err != nil {
		h.Logger.Error("Error on waiting command: " + err.Error())
		return 1, err
	}

	ans = hostCommand.ProcessState.ExitCode()

	h.Logger.DebugC(h.Logger.Aurora.Bold(
		h.Logger.Aurora.BrightYellow(
			fmt.Sprintf(":station: Exiting [%d]", ans))))

	return ans, nil
}
