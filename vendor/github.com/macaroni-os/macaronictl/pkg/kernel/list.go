/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernel

import (
	"fmt"

	"github.com/macaroni-os/macaronictl/pkg/anise"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func AvailableExtraModules(kernelBranch, kernelType string, installed bool,
	config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	ans := &specs.StonesPack{
		Stones: []*specs.Stone{},
	}

	aniseBin := utils.TryResolveBinaryAbsPath("anise")
	args := []string{
		aniseBin, "search", "-a", "kernel_module",
		"-o", "json", "--label", "kernel.type",
	}

	if kernelBranch != "" {
		var category string

		switch kernelType {
		case "zen":
			category = "kernel-zen-" + kernelBranch
		default:
			category = "kernel-" + kernelBranch
		}

		args = append(args, []string{
			"--category", category,
		}...)
	}

	if installed {
		args = append(args, "--installed")
	}

	stones, err := anise.SearchStones(args)
	if err != nil {
		return ans, err
	}

	// Filter for modules with the same kernel type.
	if kernelType != "" {
		for idx := range stones.Stones {
			pkgKernelType := stones.Stones[idx].GetLabelValue("kernel.type")
			if pkgKernelType == kernelType {
				ans.Stones = append(ans.Stones, stones.Stones[idx])
				// We manage existing and old packages without kernel.type label
				// as vanilla type.
			} else if pkgKernelType == "" && kernelType == "vanilla" {
				ans.Stones = append(ans.Stones, stones.Stones[idx])
			}
		}

	} else {
		ans = stones
	}

	return ans, nil
}

func AvailableKernels(config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	aniseBin := utils.TryResolveBinaryAbsPath("anise")
	args := []string{
		aniseBin, "search", "-a", "kernel",
		"-o", "json",
	}

	return anise.SearchStones(args)
}

func InstalledKernels(config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	aniseBin := utils.TryResolveBinaryAbsPath("anise")
	args := []string{
		aniseBin, "search", "-a", "kernel",
		"-o", "json", "--installed",
	}

	return anise.SearchStones(args)
}

func ParseKernelAnnotations(s *specs.Stone) (*specs.KernelAnnotation, error) {
	ans := &specs.KernelAnnotation{
		EoL:      "",
		Lts:      false,
		Released: "",
		Suffix:   "",
		Type:     "",
	}

	fieldsI, ok := s.Annotations["kernel"]
	if !ok {
		return nil, fmt.Errorf("[%s/%s] No kernel annotation key found",
			s.Category, s.Name)
	}

	fields, ok := fieldsI.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("[%s/%s] Error on cast annotations fields",
			s.Category, s.Name)
	}

	// Get eol
	if val, ok := fields["eol"]; ok {
		ans.EoL, _ = val.(string)
	}

	// Get lts
	if val, ok := fields["lts"]; ok {
		ans.Lts, _ = val.(bool)
	}

	// Get released
	if val, ok := fields["released"]; ok {
		ans.Released, _ = val.(string)
	}

	// Get suffix
	if val, ok := fields["suffix"]; ok {
		ans.Suffix, _ = val.(string)
	}

	// Get type
	if val, ok := fields["type"]; ok {
		ans.Type, _ = val.(string)
	}

	return ans, nil
}
