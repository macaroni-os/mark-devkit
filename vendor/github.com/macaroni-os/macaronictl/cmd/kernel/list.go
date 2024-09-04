/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdkernel

import (
	"fmt"
	"os"

	"github.com/macaroni-os/macaronictl/pkg/kernel"
	kernelspecs "github.com/macaroni-os/macaronictl/pkg/kernel/specs"
	"github.com/macaroni-os/macaronictl/pkg/profile"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewListcommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List available kernels on system.",
		Long: `Shows kernels available in your system

$ macaronictl kernel list

`,
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			bootDir, _ := cmd.Flags().GetString("bootDir")
			kernelProfilesDir, _ := cmd.Flags().GetString("kernel-profiles-dir")

			types := []kernelspecs.KernelType{}
			if kernelProfilesDir != "" {
				types, _ = profile.LoadKernelProfiles(kernelProfilesDir)
			} else {
				types, _ = profile.LoadKernelProfiles(config.GetKernelProfilesDir())
			}
			if len(types) == 0 {
				types = profile.GetDefaultKernelProfiles()
			}

			bootFiles, err := kernel.ReadBootDir(bootDir, types)
			if err != nil {
				fmt.Println("Error on read boot directory: " + err.Error())
				os.Exit(1)
			}

			if jsonOutput {
				fmt.Println(bootFiles)
			} else {

				if len(bootFiles.Files) == 0 {
					fmt.Println("No kernel files available.")
					os.Exit(0)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Kernel",
					"Kernel Version",
					"Type",
					"Has Initrd",
					"Has Kernel Image",
					"Has bzImage,Initrd links",
				})

				for _, kf := range bootFiles.Files {

					hasInitrd := false
					hasKernel := false
					hasLinks := false

					row := []string{
						kf.Type.GetSuffix(),
					}

					version := ""
					if kf.Initrd != nil {
						version = kf.Initrd.GetVersion()
						hasInitrd = true
					}

					if kf.Kernel != nil {
						version = kf.Kernel.GetVersion()
						hasKernel = true
					}

					if hasKernel && hasInitrd &&
						bootFiles.BzImageLink != "" &&
						bootFiles.InitrdLink != "" &&
						kf.Kernel.GetFilename() == bootFiles.BzImageLink &&
						kf.Initrd.GetFilename() == bootFiles.InitrdLink {
						hasLinks = true
					}

					row = append(row, []string{
						version,
						kf.Type.GetType(),
						fmt.Sprintf("%v", hasInitrd),
						fmt.Sprintf("%v", hasKernel),
						fmt.Sprintf("%v", hasLinks),
					}...)

					table.Append(row)
				}

				table.Render()
			}
		},
	}

	flags := c.Flags()
	flags.Bool("json", false, "JSON output")
	flags.String("bootdir", "/boot", "Directory where analyze kernel files.")
	flags.String("kernel-profiles-dir", "",
		"Specify the directory where read the kernel types profiles supported.")

	return c
}
