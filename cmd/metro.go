/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmdmetro "github.com/macaroni-os/mark-devkit/cmd/metro"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func metroCmdCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "metro",
		Aliases: []string{"m", "me"},
		Short:   "Build stages commands.",
		Long:    `Commands for build and manage M.A.R.K. stages.`,
	}

	cmd.AddCommand(
		cmdmetro.RunJobCommand(config),
	)

	return cmd
}
