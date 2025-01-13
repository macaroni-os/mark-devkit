/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/specs"
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
				return nil, fmt.Errorf("unsupported kind of tranform %s for atom %s",
					transform.Kind, atom.Name)
			}

			ans[versions[idx]] = v
		}
	}

	return &ans, nil
}
