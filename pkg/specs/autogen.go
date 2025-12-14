/*
Copyright © 2024-2025 Macaroni OS Linux
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
	ans := false
	for _, def := range a.Definitions {
		if def.Generator == GeneratorBuiltinGitub {
			ans = true
			break
		}
	}
	return ans
}

func (a *AutogenSpec) HasExtensionsDefs() bool {
	ans := false
	for _, def := range a.Definitions {
		if len(def.Extensions) > 0 {
			ans = true
			break
		}
	}
	return ans
}

func (a *AutogenSpec) Prepare() {
	for idx := range a.Definitions {
		if len(a.Definitions[idx].Packages) > 0 {
			for j := range a.Definitions[idx].Packages {
				for pname, atom := range a.Definitions[idx].Packages[j] {
					if atom == nil {
						// This is possible when there options defined
						a.Definitions[idx].Packages[j][pname] = NewAutogenAtom(pname)
					} else {
						if a.Definitions[idx].Packages[j][pname].Name == "" {
							a.Definitions[idx].Packages[j][pname].Name = pname
						}
					}
				}
			}
		}
	}
}

func (e *AutogenExtension) GetName() string               { return e.Name }
func (e *AutogenExtension) GetOptions() map[string]string { return e.Options }
func (e *AutogenExtension) Clone() *AutogenExtension {
	ans := &AutogenExtension{
		Name:    e.Name,
		Options: make(map[string]string, 0),
	}
	for k, v := range e.Options {
		ans.Options[k] = v
	}
	return ans
}

func NewAutogenAtom(name string) *AutogenAtom {
	return &AutogenAtom{
		Name:       name,
		Vars:       make(map[string]interface{}, 0),
		Selector:   []string{},
		Transforms: []*AutogenTransform{},
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

func (a *AutogenAtom) HasExcludes() bool {
	return len(a.Excludes) > 0
}

func (a *AutogenAtom) HasExtensions() bool {
	return len(a.Extensions) > 0
}

func (a *AutogenAtom) HasRevision() bool {
	return a.Revision != nil && *a.Revision > 0
}

func (a *AutogenAtom) String() string {
	data, _ := yaml.Marshal(a)
	return string(data)
}

func (a *AutogenAtom) GithubIgnoreTags() bool {
	if a.Github != nil && a.Github.IgnoreTags != nil {
		return *a.Github.IgnoreTags
	}
	return false
}

func (a *AutogenAtom) GetCategory(def *AutogenAtom) string {
	if a.Category != "" {
		return a.Category
	}
	if def != nil {
		return def.Category
	}
	return ""
}

func (a *AutogenAtom) GetTemplate(def *AutogenAtom) string {
	if a.Template != "" {
		return a.Template
	}

	if def != nil && def.Template != "" {
		return def.Template
	}

	return fmt.Sprintf("templates/%s.tmpl", a.Name)
}

func (a *AutogenAtom) Merge(atom *AutogenAtom) *AutogenAtom {
	ans := a.Clone()

	if atom.Name != "" {
		ans.Name = atom.Name
	}
	if atom.Tarball != "" {
		ans.Tarball = atom.Tarball
	}
	if atom.Revision != nil {
		rev := *atom.Revision
		ans.Revision = &rev
	}
	if atom.Github != nil {
		if a.Github == nil {
			ans.Github = atom.Github
		} else {
			if atom.Github.User != "" {
				ans.Github.User = atom.Github.User
			}
			if atom.Github.Repo != "" {
				ans.Github.Repo = atom.Github.Repo
			}
			if atom.Github.Query != "" {
				ans.Github.Query = atom.Github.Query
			}
			if atom.Github.PerPage != nil {
				ans.Github.PerPage = atom.Github.PerPage
			}
			if atom.Github.Page != nil {
				ans.Github.Page = atom.Github.Page
			}
			if atom.Github.NumPages != nil {
				ans.Github.NumPages = atom.Github.NumPages
			}
			if atom.Github.IgnoreTags != nil {
				ans.Github.IgnoreTags = atom.Github.IgnoreTags
			}
			if atom.Github.Match != "" {
				ans.Github.Match = atom.Github.Match
			}
		}
	}

	if atom.Dir != nil {
		if ans.Dir == nil {
			ans.Dir = atom.Dir
		} else {
			if atom.Dir.Url != "" {
				ans.Dir.Url = atom.Dir.Url
			}
			if atom.Dir.Matcher != "" {
				ans.Dir.Matcher = atom.Dir.Matcher
			}
			if atom.Dir.ExcludesMatcher != "" {
				ans.Dir.ExcludesMatcher = atom.Dir.ExcludesMatcher
			}
		}
	}

	if atom.Json != nil {
		if ans.Json == nil {
			ans.Json = atom.Json
		} else {
			if atom.Json.Url != "" {
				ans.Json.Url = atom.Json.Url
			}
			if atom.Json.Method != "" {
				ans.Json.Method = atom.Json.Method
			}
			if atom.Json.FilterVersion != "" {
				ans.Json.FilterVersion = atom.Json.FilterVersion
			}
			if atom.Json.FilterSrcUri != "" {
				ans.Json.FilterSrcUri = atom.Json.FilterSrcUri
			}
			if atom.Json.Exclude != "" {
				ans.Json.Exclude = atom.Json.Exclude
			}
			if len(atom.Json.Params) > 0 {
				for k, v := range atom.Json.Params {
					ans.Json.Params[k] = v
				}
			}
			if len(atom.Json.MapFilterVars) > 0 {
				for k, v := range atom.Json.MapFilterVars {
					ans.Json.MapFilterVars[k] = v
				}
			}
		}
	}

	if atom.Python != nil {
		if ans.Python == nil {
			ans.Python = atom.Python
		} else {
			if atom.Python.PythonCompat != "" {
				ans.Python.PythonCompat = atom.Python.PythonCompat
			}
			if atom.Python.PypiName != "" {
				ans.Python.PypiName = atom.Python.PypiName
			}
			if atom.Python.PythonRequiresIgnore != "" {
				ans.Python.PythonRequiresIgnore = atom.Python.PythonRequiresIgnore
			}
			if len(atom.Python.Pydeps) > 0 {
				for k, v := range atom.Python.Pydeps {
					ans.Python.Pydeps[k] = v
				}
			}
			if len(atom.Python.DepsIgnore) > 0 {
				for _, d := range atom.Python.DepsIgnore {
					ans.Python.DepsIgnore = append(ans.Python.DepsIgnore, d)
				}
			}
		}
	}

	if len(ans.Vars) > 0 && len(atom.Vars) > 0 {
		for k, v := range atom.Vars {
			ans.Vars[k] = v
		}

	} else if len(atom.Vars) > 0 {
		ans.Vars = atom.Vars
	}

	if atom.Category != "" {
		ans.Category = atom.Category
	}
	if atom.Template != "" {
		ans.Template = atom.Template
	}
	if atom.FilesDir != "" {
		ans.FilesDir = atom.FilesDir
	}

	if atom.HasAssets() {
		ans.Assets = atom.Assets
	}
	if atom.HasTransforms() {
		ans.Transforms = atom.Transforms
	}
	if atom.HasSelector() {
		ans.Selector = atom.Selector
	}
	if atom.HasExcludes() {
		ans.Excludes = atom.Excludes
	}

	if atom.IgnoreArtefacts != nil {
		ans.IgnoreArtefacts = atom.IgnoreArtefacts
	}

	if len(atom.Extensions) > 0 {
		for _, e := range atom.Extensions {
			present := false
			for idx := range ans.Extensions {
				if ans.Extensions[idx] == e {
					present = true
					break
				}
			}
			if !present {
				ans.Extensions = append(ans.Extensions, e)
			}
		}
	}

	return ans
}

func (a *AutogenAtom) Clone() *AutogenAtom {
	ans := &AutogenAtom{
		Name:            a.Name,
		Tarball:         a.Tarball,
		Vars:            make(map[string]interface{}, 0),
		Category:        a.Category,
		IgnoreArtefacts: a.IgnoreArtefacts,
		Template:        a.Template,
		FilesDir:        a.FilesDir,
		Transforms:      a.Transforms,
		Excludes:        a.Excludes,
		Selector:        a.Selector,
		Assets:          a.Assets,
		Extensions:      []string{},
	}

	if len(a.Vars) > 0 {
		for k, v := range a.Vars {
			ans.Vars[k] = v
		}
	}

	if a.Revision != nil {
		rev := *a.Revision
		ans.Revision = &rev
	}

	if a.Github != nil {
		ans.Github = &AutogenGithubProps{
			User:       a.Github.User,
			Repo:       a.Github.Repo,
			Query:      a.Github.Query,
			PerPage:    a.Github.PerPage,
			Page:       a.Github.Page,
			NumPages:   a.Github.NumPages,
			Match:      a.Github.Match,
			IgnoreTags: a.Github.IgnoreTags,
		}
	}

	if a.Dir != nil {
		ans.Dir = &AutogenDirlistingProps{
			Url:             a.Dir.Url,
			Matcher:         a.Dir.Matcher,
			ExcludesMatcher: a.Dir.ExcludesMatcher,
		}
	}

	if a.Json != nil {
		ans.Json = &AutogenJsonProps{
			Url:           a.Json.Url,
			Method:        a.Json.Method,
			FilterVersion: a.Json.FilterVersion,
			FilterSrcUri:  a.Json.FilterSrcUri,
			Exclude:       a.Json.Exclude,
			Params:        make(map[string]string, 0),
			MapFilterVars: make(map[string]string, 0),
		}

		if len(a.Json.Params) > 0 {
			for k, v := range a.Json.Params {
				ans.Json.Params[k] = v
			}
		}

		if len(a.Json.MapFilterVars) > 0 {
			for k, v := range a.Json.MapFilterVars {
				ans.Json.MapFilterVars[k] = v
			}
		}
	}

	if a.Python != nil {
		ans.Python = &AutogenPythonOpts{
			PythonCompat:         a.Python.PythonCompat,
			PythonRequiresIgnore: a.Python.PythonRequiresIgnore,
			DepsIgnore:           []string{},
			PypiName:             a.Python.PypiName,
			Pydeps:               make(map[string][]string, 0),
		}

		if len(a.Python.DepsIgnore) > 0 {
			for _, d := range a.Python.DepsIgnore {
				ans.Python.DepsIgnore = append(ans.Python.DepsIgnore, d)
			}
		}

		if len(a.Python.Pydeps) > 0 {
			for k, v := range a.Python.Pydeps {
				ans.Python.Pydeps[k] = v
			}
		}

	}

	if len(a.Extensions) > 0 {
		for _, e := range a.Extensions {
			ans.Extensions = append(ans.Extensions, e)
		}
	}

	return ans
}

func (a *AutogenArtefact) IsLocal() bool {
	return a.Local != nil && *a.Local
}

func (a *AutogenDefinition) GetExtensionOptions(e string) (*AutogenExtension, error) {
	ext, ok := a.Extensions[e]
	if ok && ext != nil {
		ans := ext.Clone()
		if ans.Name == "" {
			ans.Name = e
		}
		return ans, nil
	}
	return nil, fmt.Errorf("Extension %s without options", e)
}
