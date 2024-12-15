/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmdkit "github.com/macaroni-os/mark-devkit/cmd/kit"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func kitCmdCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "kit",
		Aliases: []string{"k"},
		Short:   "Kit commands.",
		Long:    `Executes Kits commands.`,
	}

	cmd.AddCommand(
		cmdkit.KitCloneCommand(config),
		cmdkit.KitMergeCommand(config),
	)

	return cmd
}
