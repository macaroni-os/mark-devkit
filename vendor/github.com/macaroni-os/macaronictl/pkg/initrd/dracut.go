/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package initrd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	kernelspecs "github.com/macaroni-os/macaronictl/pkg/kernel/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type DracutBuilder struct {
	DryRun bool
	Args   string
}

func NewDracutBuilder(args string, dryRun bool) *DracutBuilder {
	ans := &DracutBuilder{
		Args:   args,
		DryRun: dryRun,
	}

	if os.Getenv("MACARONICTL_DRACUT_ARGS") != "" {
		ans.Args = os.Getenv("MACARONICTL_DRACUT_ARGS")
	}

	return ans
}

func (d *DracutBuilder) Build(kf *kernelspecs.KernelFiles, bootDir string) error {
	if kf == nil || kf.Kernel == nil || kf.Type == nil {
		return errors.New("Invalid kernel file")
	}

	kverstr := kf.Kernel.GetVersion()
	if kf.Kernel.GetSuffix() != "" {
		kverstr += "-" + kf.Kernel.GetSuffix()
	}

	// Convert args in array
	args := strings.Split(d.Args, " ")

	initrd := kf.Initrd
	if kf.Initrd == nil {
		initrd = kernelspecs.NewInitrdImage()
		initrd.SetPrefix(kf.Type.GetInitrdPrefixSanitized())
		initrd.SetVersion(kf.Kernel.GetVersion())
		initrd.SetSuffix(kf.Type.GetSuffix())
		initrd.SetKernelType(kf.Kernel.GetType())
		initrd.SetArch(kf.Kernel.GetArch())
	}

	kf.Initrd = initrd

	if d.DryRun {
		fmt.Println("[dry-run mode] command: dracut " + d.Args)
		return nil
	}

	initrdFile := filepath.Join(bootDir, initrd.GenerateFilename())

	args = append(args, []string{
		"--kver", kverstr, initrdFile,
	}...)

	fmt.Print(fmt.Sprintf("Creating initrd image %s...", initrdFile))

	dracut := utils.TryResolveBinaryAbsPath("dracut")
	dracutCommand := exec.Command(dracut, args...)
	dracutCommand.Stdout = os.Stdout
	dracutCommand.Stderr = os.Stderr

	err := dracutCommand.Start()
	if err != nil {
		return errors.New("Error on start dracut command: " + err.Error())
	}

	err = dracutCommand.Wait()
	if err != nil {
		return errors.New("Error on waiting dracut command: " + err.Error())
	}

	if dracutCommand.ProcessState.ExitCode() != 0 {
		return errors.New(
			fmt.Sprintf("dracut command exiting with %d",
				dracutCommand.ProcessState.ExitCode()))
	}

	fmt.Println("DONE")

	return nil
}
