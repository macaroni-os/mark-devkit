/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package metro

import (
	"os"

	specs "github.com/macaroni-os/mark-devkit/pkg/specs"
)

func (m *Metro) Load(f string) (*specs.MetroSpec, error) {
	ans := specs.NewMetroSpec()

	// Read the main specification file
	content, err := os.ReadFile(f)
	if err != nil {
		return ans, err
	}

	err = ans.LoadYaml(content, f)
	if err != nil {
		return ans, err
	}

	return ans, nil
}
