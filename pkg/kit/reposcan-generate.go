/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func RunReposcanGenerate(sourceDir, kitName, kitBranch, targetFile string,
	eclassDirs []string, concurrency int, verbose bool) error {
	anisePcBin := utils.TryResolveBinaryAbsPath("anise-portage-converter")

	args := []string{
		anisePcBin, "reposcan-generate",
		"--kit", kitName,
		"--branch", kitBranch,
		sourceDir,
		"--concurrency", fmt.Sprintf("%d", concurrency),
		"-o", "file",
		"-f", targetFile,
	}

	for _, dir := range eclassDirs {
		args = append(args,
			[]string{
				"--eclass-dir", dir,
			}...)
	}

	cmd := exec.Command(args[0], args[1:]...)
	if verbose {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("anise-portage-converter exiting with %s.",
			cmd.ProcessState.ExitCode())
	}

	return nil
}

func (m *MergeBot) GenerateKitCacheFile(sourceDir, kitName, kitBranch, targetFile string,
	eclassDirs []string, concurrency int, verbose bool) error {

	m.Logger.Debug(fmt.Sprintf("Generating kit-cache file for kit %s...",
		kitName))

	return RunReposcanGenerate(sourceDir, kitName, kitBranch, targetFile,
		eclassDirs, concurrency, verbose)
}
