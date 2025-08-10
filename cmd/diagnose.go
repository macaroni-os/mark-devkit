/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmddiag "github.com/macaroni-os/mark-devkit/cmd/diagnose"
	specs "github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/spf13/cobra"
)

func diagnoseCmdCommand(config *specs.MarkDevkitConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "diagnose",
		Aliases: []string{"di", "d"},
		Short:   "Diagnose commands.",
		Long:    `Executes diagnose commands to validate and check operations.`,
	}

	cmd.AddCommand(
		cmddiag.DiagnoseJobCommand(config),
	)

	return cmd
}
