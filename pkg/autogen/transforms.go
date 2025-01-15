/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

func (a *AutogenBot) transformsVersions(atom *specs.AutogenAtom, versions []string) (*map[string]string, error) {
	ans := make(map[string]string, 0)
	var v string

	for _, transform := range atom.Transforms {
		for idx := range versions {
			if elabVer, ok := ans[versions[idx]]; ok {
				v = elabVer
			} else {
				v = versions[idx]
			}
			switch transform.Kind {
			case "string":
				v = strings.ReplaceAll(v, transform.Match, transform.Replace)
			default:
				return nil, fmt.Errorf("unsupported kind of transform %s for atom %s",
					transform.Kind, atom.Name)
			}

			ans[versions[idx]] = v
		}
	}

	return &ans, nil
}

func (a *AutogenBot) sortVersions(atom, def *specs.AutogenAtom, versions []string) ([]string, error) {
	// In order to avoid issues with go-version on parse particular versions
	// I prefer to use GentooPackage that already support different versions and sorting.

	ans := []string{}
	pkgs := []gentoo.GentooPackage{}

	for idx := range versions {
		gp, err := gentoo.ParsePackageStr(fmt.Sprintf("%s/%s-%s",
			atom.GetCategory(def),
			atom.Name,
			versions[idx],
		))
		if err != nil {
			return ans, err
		}

		pkgs = append(pkgs, *gp)
	}

	sort.Sort(sort.Reverse(gentoo.GentooPackageSorter(pkgs)))

	for idx := range pkgs {
		ans = append(ans, pkgs[idx].GetPVR())
	}

	return ans, nil
}
