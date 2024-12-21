/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

type MergeKit struct {
	Sources []*ReposcanKit `yaml:"sources,omitempty" json:"sources,omitempty"`
	Target  MergeKitTarget `yaml:"target,omitempty" json:"target,omitempty"`

	File string `yaml:"-" json:"-"`
}

type MergeKitTarget struct {
	Name   string `yaml:"name" json:"name"`
	Url    string `yaml:"url,omitempty" json:"url,omitempty"`
	Branch string `yaml:"branch" json:"branch"`

	Eclasses *MergeKitEclasses `yaml:"eclasses,omitempty" json:"eclasses,omitempty"`
	Metadata *MergeKitMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Fixups   *MergeKitFixups   `yaml:"fixups,omitempty" json:"fixups,omitempty"`

	AtomDefaults *MergeKitAtom   `yaml:"atoms_defaults,omitempty" json:"atoms_defaults,omitempty"`
	Atoms        []*MergeKitAtom `yaml:"atoms,omitempty" json:"atoms,omitempty"`
}

type MergeKitMetadata struct {
	LayoutMasters string   `yaml:"layout_masters,omitempty" json:"layout_masters,omitempty"`
	Aliases       []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
}

type MergeKitEclasses struct {
	Include map[string][]string `yaml:"include,omitempty" json:"include,omitempty"`
}

type MergeKitAtom struct {
	Package     string   `yaml:"pkg,omitempty" json:"pkg,omitempty"`
	MaxVersions *int     `yaml:"max_versions,omitempty" json:"max_versions,omitempty"`
	Conditions  []string `yaml:"conditions,omitempty" json:"conditions,omitempty"`
}

type MergeKitFixups struct {
	Include []*MergeKitFixupInclude `yaml:"include,omitempty" json:"include,omitempty"`
}

type MergeKitFixupInclude struct {
	Dir  string `yaml:"dir,omitempty" json:"dir,omitempty"`
	To   string `yaml:"to,omitempty" json:"to,omitempty"`
	File string `yaml:"file,omitempty" json:"file,omitempty"`
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}
