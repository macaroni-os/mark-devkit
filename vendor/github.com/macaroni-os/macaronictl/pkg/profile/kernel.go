/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package profile

import (
	"io/ioutil"
	"path"
	"regexp"

	kernelspecs "github.com/macaroni-os/macaronictl/pkg/kernel/specs"
)

func GetDefaultKernelProfiles() []kernelspecs.KernelType {
	return []kernelspecs.KernelType{
		{
			Name:     "Sabayon",
			Suffix:   "sabayon",
			Type:     "genkernel",
			WithArch: true,
		},
		{
			Name:     "Macaroni",
			Suffix:   "macaroni",
			Type:     "vanilla",
			WithArch: true,
		},
		{
			Name:     "Macaroni Zen Kernel",
			Suffix:   "macaroni",
			Type:     "zen",
			WithArch: true,
		},
	}
}

func LoadKernelProfiles(dir string) ([]kernelspecs.KernelType, error) {
	ans := []kernelspecs.KernelType{}

	var regexRepo = regexp.MustCompile(`.yml$|.yaml$`)
	var err error

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ans, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !regexRepo.MatchString(file.Name()) {
			continue
		}

		content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			// TODO: integrate warning logger
			continue
		}

		ktype, err := kernelspecs.KernelTypeFromYaml(content)
		if err != nil {
			continue
		}

		if ktype.GetType() != "" {
			ans = append(ans, *ktype)
		}
	}

	return ans, nil
}
