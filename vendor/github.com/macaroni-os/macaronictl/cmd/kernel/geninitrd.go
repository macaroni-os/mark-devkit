/*
	Copyright © 2021 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmdkernel

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macaroni-os/macaronictl/pkg/initrd"
	"github.com/macaroni-os/macaronictl/pkg/kernel"
	kernelspecs "github.com/macaroni-os/macaronictl/pkg/kernel/specs"
	"github.com/macaroni-os/macaronictl/pkg/logger"
	"github.com/macaroni-os/macaronictl/pkg/profile"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"

	"github.com/spf13/cobra"
)

func setFilesLinks(kf *kernelspecs.KernelFiles, bootDir, release string) error {

	log := logger.GetDefaultLogger()

	// Ignoring errors
	os.Remove(filepath.Join(bootDir, "bzImage"))
	os.Remove(filepath.Join(bootDir, "Initrd"))

	err := os.Chdir(bootDir)
	if err != nil {
		return err
	}

	log.DebugC("Creating link bzImage to", kf.Kernel.GetFilename())
	err = os.Symlink(kf.Kernel.GetFilename(), filepath.Join(bootDir, "bzImage"))
	if err != nil {
		return err
	}

	if kf.Initrd == nil {
		fmt.Println(fmt.Sprintf(
			"WARN: No initrd image found for kernel %s. Initrd symbolic link is not created.",
			kf.Kernel.GetVersion(),
		))

	} else {
		log.DebugC("Creating link Initrd to", kf.Initrd.GenerateFilename())
		err = os.Symlink(kf.Initrd.GenerateFilename(), filepath.Join(bootDir, "Initrd"))
		if err != nil {
			return err
		}
	}

	return nil
}

func NewGeninitrdCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "geninitrd",
		Aliases: []string{"gi"},
		Short:   "Generate initrd image and set default kernel/initrd links.",
		Long: `Rebuild Dracut initrd images.

$> # Generate all initrd images of the kernels available on boot dir.
$> macaronictl kernel geninitrd --all

$> # Generate all initrd images of the kernels available on boot dir
$> # and set the bzImage, Initrd links to one of the kernel available
$> # if not present or to the next release of the same kernel after the
$> # upgrade.
$> macaronictl kernel geninitrd --all --set-links

$> # Generate all initrd images of the kernels available on boot dir
$> # and set the bzImage, Initrd links to one of the kernel available
$> # if not present or to the next release of the same kernel after the
$> # upgrade. In addition, it purges old initrd images and update grub.cfg.
$> macaronictl kernel geninitrd --all --set-links --purge --grub

$> # Just show what dracut commands will be executed for every initrd images.
$> macaronictl kernel geninitrd --all --dry-run

$> # Generate the initrd image for the kernel 5.10.42
$> macaronictl kernel geninitrd --version 5.10.42

$> # Generate the initrd image for the kernel 5.10.42 and kernel type vanilla.
$> macaronictl kernel geninitrd --version 5.10.42 --ktype vanilla

$> # Generate the initrd image for the kernel 5.10.42 and kernel type vanilla
$> # and set the links bzImage, Initrd to the selected kernel/initrd.
$> macaronictl kernel geninitrd --version 5.10.42 --ktype vanilla

`,
		PreRun: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			version, _ := cmd.Flags().GetString("version")
			if !all && version == "" {
				fmt.Println("You need to use --all or --version")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			bootDir, _ := cmd.Flags().GetString("bootDir")
			all, _ := cmd.Flags().GetBool("all")
			setLinks, _ := cmd.Flags().GetBool("set-links")
			version, _ := cmd.Flags().GetString("version")
			ktype, _ := cmd.Flags().GetString("ktype")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			dracutOpts, _ := cmd.Flags().GetString("dracut-opts")
			purge, _ := cmd.Flags().GetBool("purge")
			grub, _ := cmd.Flags().GetBool("grub")
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

			release, err := utils.OsRelease()
			if err != nil {
				fmt.Println("Error on retrieve os release: " + err.Error())
				os.Exit(1)
			}

			// TODO: default dracut options will be read from configuration.

			defaultDracutOpts :=
				"-H -q -f -o systemd -o systemd-initrd -o systemd-networkd -o dracut-systemd"
			if dracutOpts != "" {
				defaultDracutOpts = dracutOpts
			}
			dracutBuilder := initrd.NewDracutBuilder(defaultDracutOpts, dryRun)

			if all {
				for idx, f := range bootFiles.Files {
					if f.Kernel == nil {
						// Ignore initrd without kernel image.
						continue
					}

					err := dracutBuilder.Build(bootFiles.Files[idx], bootFiles.Dir)
					if err != nil {
						fmt.Println(fmt.Sprintf("Error on generate initrd image for kernel %s: %s. I go ahead.",
							f.Kernel.GetFilename(),
							err.Error(),
						))
					}
				}

				var kf *kernelspecs.KernelFiles

				if setLinks && !bootFiles.BzImageLinkExistingKernel() {

					// Retrieve the kernel matching for type, prefix and suffix
					kf = bootFiles.RetrieveBzImageSelectedKernel()
					if kf == nil {
						for _, file := range bootFiles.Files {
							if file.Kernel != nil {
								kf = file
							}
						}

						if kf != nil {
							fmt.Println(fmt.Sprintf(
								"No valid links found. I set links to kernel %s.",
								kf.Kernel.GetVersion(),
							))
						}
					}

					if kf != nil {
						err := setFilesLinks(kf, bootFiles.Dir, release)
						if err != nil {
							fmt.Println(fmt.Sprintf("Error on set links for kernel %s: %s",
								kf.Kernel.GetVersion(),
								err.Error(),
							))
						}
					}
				}

			} else {

				file, err := bootFiles.GetFile(version, ktype)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				err = dracutBuilder.Build(file, bootFiles.Dir)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on generate initrd image for kernel %s: %s. I go ahead.",
						file.Kernel.GetFilename(),
						err.Error(),
					))
				}

				if setLinks {

					err := setFilesLinks(file, bootFiles.Dir, release)
					if err != nil {
						fmt.Println(fmt.Sprintf("Error on set links for kernel %s: %s",
							file.Kernel.GetVersion(),
							err.Error(),
						))
					}
				}

			}

			// Purge orphan initrd
			if purge {
				err = bootFiles.PurgeOrphanInitrdImages()
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on purge orphan initrd images: %s", err.Error()))
				}
			}

			// Update grub config
			if grub {
				err = kernel.GrubMkconfig(filepath.Join(bootFiles.Dir, "/grub/grub.cfg"), dryRun)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on update grub.cfg: %s", err.Error()))
					// TODO: We need ignore it?
					os.Exit(1)
				}
			}

		},
	}

	flags := c.Flags()
	flags.Bool("all", false, "Rebuild all images with kernel.")
	flags.Bool("dry-run", false, "Dry run commands.")
	flags.Bool("set-links", false, "Set bzImage and Initrd links for the selected kernel or update links of the upgraded kernel.")
	flags.Bool("purge", false, "Clean orphan initrd images without kernel.")
	flags.Bool("grub", false, "Update grub.cfg.")
	flags.String("bootdir", "/boot", "Directory where analyze kernel files.")
	flags.String("version", "", "Specify the kernel version of the initrd image to build.")
	flags.String("ktype", "", "Specify the kernel type of the initrd image to build.")
	flags.String("dracut-opts", "",
		`Override the default dracut options used on the initrd image generation.
Set the MACARONICTL_DRACUT_ARGS env in alternative.`)
	flags.String("kernel-profiles-dir", "",
		"Specify the directory where read the kernel types profiles supported.")

	return c
}
