/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

const (
	CacheDataVersion = "1.0.6"
)

type RepoScanSpec struct {
	CacheDataVersion string                  `json:"cache_data_version" yaml:"cache_data_version"`
	Atoms            map[string]RepoScanAtom `json:"atoms" yaml:"atoms"`
	MetadataErrors   map[string]RepoScanAtom `json:"metadata_errors,omitempty" yaml:"metadata_errors,omitempty"`

	File string `json:"-"`
}

type RepoScanAtom struct {
	Atom string `json:"atom,omitempty" yaml:"atom,omitempty"`

	Category string     `json:"category,omitempty" yaml:"category,omitempty"`
	Package  string     `json:"package,omitempty" yaml:"package,omitempty"`
	Revision string     `json:"revision,omitempty" yaml:"revision,omitempty"`
	CatPkg   string     `json:"catpkg,omitempty" yaml:"catpkg,omitempty"`
	Eclasses [][]string `json:"eclasses,omitempty" yaml:"eclasses,omitempty"`

	Kit    string `json:"kit,omitempty" yaml:"kit,omitempty"`
	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`

	// Relations contains the list of the keys defined on
	// relations_by_kind. The values could be RDEPEND, DEPEND, BDEPEND
	Relations       []string            `json:"relations,omitempty" yaml:"relations,omitempty"`
	RelationsByKind map[string][]string `json:"relations_by_kind,omitempty" yaml:"relations_by_kind,omitempty"`

	// Metadata contains ebuild variables.
	// Ex: SLOT, SRC_URI, HOMEPAGE, etc.
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	MetadataOut string            `json:"metadata_out,omitempty" yaml:"metadata_out,omitempty"`

	ManifestMd5 string `json:"manifest_md5,omitempty" yaml:"manifest_md5,omitempty"`
	Md5         string `json:"md5,omitempty" yaml:"md5,omitempty"`

	// Fields present on failure
	Status string `json:"status,omitempty" yaml:"status,omitempty"`
	Output string `json:"output,omitempty" yaml:"output,omitempty"`

	Files []RepoScanFile `json:"files,omitempty" yaml:"files,omitempty"`
}

type RepoScanFile struct {
	SrcUri []string          `json:"src_uri"`
	Size   string            `json:"size"`
	Hashes map[string]string `json:"hashes"`
	Name   string            `json:"name"`
}

type ReposcanAnalysis struct {
	Kits []*ReposcanKit `yaml:"kits,omitempty" json:"kits,omitempty"`
}

type ReposcanKit struct {
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	Url        string `yaml:"url,omitempty" json:"url,omitempty"`
	Branch     string `yaml:"branch,omitempty" json:"branch,omitempty"`
	CommitSha1 string `yaml:"commit_sha1,omitempty" json:"commit_sha1,omitempty"`
	Depth      *int   `yaml:"depth,omitempty" json:"depth,omitempty"`
	Priority   *int   `yaml:"priority,omitempty" json:"priority,omitempty"`
}
