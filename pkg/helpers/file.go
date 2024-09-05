/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers

import (
	"os"
)

func EnsureDir(dir string, uid, gid int, perm os.FileMode) error {
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return err
	}
	err = os.Chown(dir, uid, gid)
	if err != nil {
		return err
	}

	return nil
}
