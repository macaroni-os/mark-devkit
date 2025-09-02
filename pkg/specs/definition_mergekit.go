/*
Copyright © 2024-2025 Macaroni OS Linux
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

	Eclasses          *MergeKitEclasses           `yaml:"eclasses,omitempty" json:"eclasses,omitempty"`
	Metadata          *MergeKitMetadata           `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	ThirdpartyMirrors []*MergeKitThirdPartyMirror `yaml:"thirdpartymirrors,omitempty" json:"thirdpartymirrors,omitempty"`
	Fixups            *MergeKitFixups             `yaml:"fixups,omitempty" json:"fixups,omitempty"`

	AtomDefaults *MergeKitAtom   `yaml:"atoms_defaults,omitempty" json:"atoms_defaults,omitempty"`
	Atoms        []*MergeKitAtom `yaml:"atoms,omitempty" json:"atoms,omitempty"`
}

type MergeKitMetadata struct {
	LayoutMasters          string   `yaml:"layout_masters,omitempty" json:"layout_masters,omitempty"`
	Aliases                []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	ManifestHashes         []string `yaml:"manifest_hashes,omitempty" json:"manifest_hashes,omitempty"`
	ManifestRequiredHashes []string `yaml:"manifest_required_hashes,omitempty" json:"manifest_required_hashes,omitempty"`
}

type MergeKitThirdPartyMirror struct {
	Alias  string        `yaml:"alias,omitempty" json:"alias,omitempty"`
	Uri    []string      `yaml:"uri,omitempty" json:"uri,omitempty"`
	Layout *MirrorLayout `yaml:"layout,omitempty" json:"layout,omitempty"`
}

type MergeKitEclasses struct {
	Include map[string][]string `yaml:"include,omitempty" json:"include,omitempty"`
}

type MergeKitAtom struct {
	Package        string   `yaml:"pkg,omitempty" json:"pkg,omitempty"`
	MaxVersions    *int     `yaml:"max_versions,omitempty" json:"max_versions,omitempty"`
	Conditions     []string `yaml:"conditions,omitempty" json:"conditions,omitempty"`
	CondIgnoreSlot *bool    `yaml:"cond_ignore_slot,omitempty" json:"cond_ignore_slot,omitempty"`
	Versions       []string `yaml:"versions,omitempty" json:"versions,omitempty"`
}

type MergeKitFixups struct {
	Include []*MergeKitFixupInclude `yaml:"include,omitempty" json:"include,omitempty"`
}

type MergeKitFixupInclude struct {
	Dir       string   `yaml:"dir,omitempty" json:"dir,omitempty"`
	To        string   `yaml:"to,omitempty" json:"to,omitempty"`
	File      string   `yaml:"file,omitempty" json:"file,omitempty"`
	Name      string   `yaml:"name,omitempty" json:"name,omitempty"`
	KeepFiles []string `yaml:"keep_files,omitempty" json:"keep_files,omitempty"`
}

type DistfilesSpec struct {
	*MergeKit       `json:"-,inline" yaml:"-,inline"`
	FallbackMirrors []*MergeKitThirdPartyMirror `json:"fallback_mirrors,omitempty" yaml:"fallback_mirrors,omitempty"`
}

type MirrorLayout struct {
	Modes []*MirrorLayoutMode `json:"modes" yaml:"modes"`
}

type MirrorLayoutMode struct {
	Type     string `json:"type" yaml:"type"`
	Hash     string `json:"hash,omitempty" yaml:"hash,omitempty"`
	HashMode string `json:"hash_mode,omitempty" yaml:"hash_mode,omitempty"`
}
