/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers

import (
	"fmt"
	"strings"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

func DecodeCondition(condition, cat, pn string) (*gentoo.GentooPackage, error) {
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
	} else if strings.HasPrefix(condition, "==") {
		condition = condition[2:]
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

	} else if strings.HasPrefix(condition, "!=") {
		condition = condition[2:]
		gcond = gentoo.PkgCondNot
	} else if strings.HasPrefix(condition, "!") {
		condition = condition[1:]
		gcond = gentoo.PkgCondNot
	}

	gpkgCond, err := gentoo.ParsePackageStr(
		fmt.Sprintf(
			"%s/%s-%s", cat, pn, condition))
	if err != nil {
		return nil, err
	}

	gpkgCond.Condition = gcond
	return gpkgCond, nil
}
