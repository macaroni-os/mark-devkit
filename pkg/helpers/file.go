/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers

import (
	"crypto/md5"
	"fmt"
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

func EnsureDirWithoutIds(dir string, perm os.FileMode) error {
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return err
	}

	return nil
}

func GetFilesMd5(f string) (string, error) {
	content, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(content)), nil
}

func CopyFile(source, target string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(target, content, 0644)
}
