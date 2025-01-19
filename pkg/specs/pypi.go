/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"gopkg.in/yaml.v3"
)

var (
	supportedPythonVersions = []string{
		"3.9",
	}
)

func NewPypiMetadata() *PypiMetadata {
	return &PypiMetadata{
		Releases: make(map[string][]*PypiReleaseFile),
	}
}

func (p *PypiMetadata) GetVersions(pythonCompat string) []string {
	ans := []string{}
	gpVersions := []*gentoo.GentooPackage{}

	if pythonCompat != "" {
		words := strings.Split(pythonCompat, " ")
		// NOTE: I consider to have always python3* together with pypi3
		//       and never only pypi3
		for _, w := range words {
			if w == "pypi3" {
				// Ignore
				continue
			}
			if w == "python3+" {
				for _, pv := range supportedPythonVersions {
					gp, _ := gentoo.ParsePackageStr(fmt.Sprintf(
						"dev-lang/python-%s", pv))
					gpVersions = append(gpVersions, gp)
				}
			} else {
				pv := strings.ReplaceAll(w, "python", "")
				gp, _ := gentoo.ParsePackageStr(fmt.Sprintf(
					"dev-lang/python-%s", pv))
				gpVersions = append(gpVersions, gp)
			}

		}
	}

	for v, rfList := range p.Releases {
		if pythonCompat != "" && len(rfList) > 0 && rfList[0].RequiresPython != "" {
			condLabels := strings.Split(rfList[0].RequiresPython, ",")

			admitted := false
			for _, cond := range condLabels {
				if strings.HasPrefix("!=", cond) {
					admitted = true
					continue
				}

				reqGp, err := helpers.DecodeCondition(
					cond, "dev-lang", "python")
				if err != nil {
					continue
				}

				for idx := range gpVersions {
					admitted, err = reqGp.Admit(gpVersions[idx])
					if err != nil {
						continue
					}
					if admitted {
						admitted = true
						break
					}
				}
			}
			if admitted {
				ans = append(ans, v)
			}

		} else {
			ans = append(ans, v)
		}
	}

	return ans
}

func (p *PypiMetadata) GetReleaseFiles(version, packagetype string) []*PypiReleaseFile {
	ans := []*PypiReleaseFile{}

	if files, present := p.Releases[version]; present {

		if packagetype == "" {
			ans = append(ans, files...)
		} else {
			for idx := range files {
				if files[idx].PackageType == packagetype {
					ans = append(ans, files[idx])
				}
			}
		}

	}

	return ans
}

func (p *PypiMetadata) GetInfo() *PypiPackageInfo { return p.Info }

func (p *PypiMetadata) Json() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

func (p *PypiMetadata) Yaml() ([]byte, error) {
	return yaml.Marshal(p)
}
