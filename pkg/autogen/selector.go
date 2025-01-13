/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package autogen

import (
	"fmt"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

func (a *AutogenBot) selectVersion(atom *specs.AutogenAtom, def *specs.AutogenAtom,
	versions []string) (string, error) {

	ans := ""
	// NOTE: We use GentooPackage.Admit to select the version
	vMap := make(map[string]*gentoo.GentooPackage, 0)

	a.Logger.Debug(fmt.Sprintf(
		":eye: [%s] Checking versions: %s",
		atom.Name, versions))

	// Prepare conditions
	conditions := atom.Selector
	if def.HasSelector() {
		conditions = append(conditions, def.Selector...)
	}
	for _, condition := range conditions {

		if condition == "" {
			return ans, fmt.Errorf("empty condition on selector for package %s", atom.Name)
		}

		origCond := condition
		var gcond gentoo.PackageCond = gentoo.PkgCondInvalid

		if strings.HasPrefix(condition, ">=") {
			condition = condition[2:]
			gcond = gentoo.PkgCondGreaterEqual
		} else if strings.HasPrefix(condition, ">") {
			condition = condition[1:]
			gcond = gentoo.PkgCondGreater
		} else if strings.HasPrefix(condition, "<=") {
			condition = condition[2:]
			gcond = gentoo.PkgCondLessEqual
		} else if strings.HasPrefix(condition, "<") {
			condition = condition[1:]
			gcond = gentoo.PkgCondLess
		} else if strings.HasPrefix(condition, "=") {
			condition = condition[1:]
			if strings.HasSuffix(condition, "*") {
				gcond = gentoo.PkgCondMatchVersion
				condition = condition[0 : len(condition)-1]
			} else {
				gcond = gentoo.PkgCondEqual
			}
		} else if strings.HasPrefix(condition, "~") {
			condition = condition[1:]
			gcond = gentoo.PkgCondAnyRevision
		} else if strings.HasPrefix(condition, "!<") {
			condition = condition[2:]
			gcond = gentoo.PkgCondNotLess
		} else if strings.HasPrefix(condition, "!>") {
			condition = condition[2:]
			gcond = gentoo.PkgCondNotGreater
		} else if strings.HasPrefix(condition, "!") {
			condition = condition[1:]
			gcond = gentoo.PkgCondNot
		}

		gpkgCond, err := gentoo.ParsePackageStr(
			fmt.Sprintf(
				"%s/%s-%s", atom.GetCategory(def), atom.Name, condition))
		if err != nil {
			return ans, err
		}

		gpkgCond.Condition = gcond
		vMap[origCond] = gpkgCond
	}

	for idx := range versions {
		gpkg, err := gentoo.ParsePackageStr(
			fmt.Sprintf(
				"%s/%s-%s", atom.GetCategory(def), atom.Name, versions[idx]))
		if err != nil {
			return ans, err
		}

		vAdmit := true

		for condition, g := range vMap {
			admitted, err := g.Admit(gpkg)
			if err != nil {
				return ans, fmt.Errorf("error on check condition %s: %s",
					condition, err.Error())
			}

			if !admitted {
				vAdmit = false
				a.Logger.Debug(fmt.Sprintf(
					":warning: [%s] Version %s skipped.",
					atom.Name, versions[idx]))
				break
			}
		}

		if vAdmit {
			ans = versions[idx]
			break
		}
	}

	if ans == "" {
		return ans, fmt.Errorf("No valid version found")
	}

	return ans, nil
}
