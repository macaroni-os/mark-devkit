/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

func OsRelease() (string, error) {
	mosReleaseFile := "/etc/macaroni/release"
	release := ""

	_, err := os.Stat(mosReleaseFile)
	if err == nil {
		content, err := ioutil.ReadFile(mosReleaseFile)
		if err != nil {
			return "", err
		}

		release = string(content)
		release = strings.ReplaceAll(release, "\n", "")
	} else if !os.IsNotExist(err) {
		return release, err
	} // else is not a macaroni os rootfs.

	return release, nil
}
