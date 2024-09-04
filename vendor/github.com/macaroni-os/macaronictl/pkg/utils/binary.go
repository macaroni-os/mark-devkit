/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package utils

import (
	"path/filepath"
)

// Try to resolve the abs path of the passed param
// Return itself if the binary is not present in the
// default paths (/sbin, /bin, /usr/sbin, /usr/bin)
func TryResolveBinaryAbsPath(b string) string {
	ans := b
	possiblePaths := []string{
		"/sbin",
		"/bin",
		"/usr/sbin",
		"/usr/bin",
	}

	for _, s := range possiblePaths {
		abs := filepath.Join(s, b)
		if Exists(abs) {
			ans = filepath.Join(abs)
			break
		}
	}

	return ans
}
