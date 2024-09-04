/*
	Copyright © 2021-2023 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmdkernel "github.com/macaroni-os/macaronictl/cmd/kernel"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/spf13/cobra"
)

func kernelCmdCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "kernel",
		Aliases: []string{"k"},
		Short:   "Manage system kernels and initrd.",
		Long:    `Manage kernels and initrd images of your system.`,
	}

	cmd.AddCommand(
		cmdkernel.NewListcommand(config),
		cmdkernel.NewAvailableCommand(config),
		cmdkernel.NewModulesCommand(config),
		cmdkernel.NewSwitchCommand(config),
		cmdkernel.NewGeninitrdCommand(config),
		cmdkernel.NewProfilesCommand(config),
	)

	return cmd
}
