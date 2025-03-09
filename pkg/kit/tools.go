/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"strings"
)

func GetPrBranchNameForEclasses(branch string) string {
	return fmt.Sprintf(
		"%s%s/%s",
		prBranchPrefix, branch, "eclasses",
	)
}

func GetPrBranchNameForMetadata(branch string) string {
	return fmt.Sprintf(
		"%s%s/%s",
		prBranchPrefix, branch, "metadata-update",
	)
}

func GetPrBranchNameForPkgBump(pkg, branch string) string {
	return fmt.Sprintf(
		"%s%s/%s-%s",
		prBranchPrefix, branch, "bump",
		strings.ReplaceAll(strings.ReplaceAll(pkg, ".", "_"),
			"/", "_"),
	)
}

func GetPrBranchNameForFixup(name, branch string) string {
	return fmt.Sprintf(
		"%s%s/%s-%s",
		prBranchPrefix, branch, "fixup-include",
		strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"),
			"/", "_"),
	)
}

func GetPrBranchNameForProfile(branch string) string {
	return fmt.Sprintf(
		"%s%s/%s",
		prBranchPrefix, branch, "profiles-update",
	)
}
