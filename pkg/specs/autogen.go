/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func NewAutogenSpec() *AutogenSpec {
	return &AutogenSpec{
		Version:     "1",
		Definitions: make(map[string]*AutogenDefinition, 0),
	}
}

func (a *AutogenSpec) LoadFile(file string) error {
	// Read specfile
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return a.LoadYaml(content, file)
}

func (a *AutogenSpec) LoadYaml(data []byte, file string) error {
	if err := yaml.Unmarshal(data, a); err != nil {
		return err
	}
	a.File = file
	return nil
}

func (a *AutogenSpec) HasGithubGenerators() bool {
	// TODO
	return true
}

func (a *AutogenSpec) Prepare() {
	for idx := range a.Definitions {
		if len(a.Definitions[idx].Packages) > 0 {
			for j := range a.Definitions[idx].Packages {
				for pname, _ := range a.Definitions[idx].Packages[j] {
					if a.Definitions[idx].Packages[j][pname].Name == "" {
						a.Definitions[idx].Packages[j][pname].Name = pname
					}
				}
			}
		}
	}
}

func (a *AutogenAtom) HasTransforms() bool {
	return len(a.Transforms) > 0
}

func (a *AutogenAtom) HasAssets() bool {
	return len(a.Assets) > 0
}

func (a *AutogenAtom) HasSelector() bool {
	return len(a.Selector) > 0
}

func (a *AutogenAtom) GetCategory(def *AutogenAtom) string {
	if a.Category != "" {
		return a.Category
	}
	return def.Category
}

func (a *AutogenAtom) GetTemplate(def *AutogenAtom) string {
	if a.Template != "" {
		return a.Template
	}

	if def.Template != "" {
		return def.Template
	}

	return fmt.Sprintf("templates/%s.tmpl", a.Name)
}

func (a *AutogenAtom) Clone() *AutogenAtom {
	ans := &AutogenAtom{
		Name:         a.Name,
		Tarball:      a.Tarball,
		Vars:         a.Vars,
		Category:     a.Category,
		PythonCompat: a.PythonCompat,
	}

	if a.Github != nil {
		ans.Github = &AutogenGithubProps{
			User:  a.Github.User,
			Repo:  a.Github.Repo,
			Query: a.Github.Query,
		}
	}

	return ans
}
