/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

// kit-info.json structure
type MetaKitInfo struct {
	KitOrder    []string                  `yaml:"kit_order,omitempty" json:"kit_order,omitempty"`
	KitSettings map[string]MetaKitSetting `yaml:"kit_settings,omitempty" json:"kit_settings,omitempty"`
	ReleaseDefs map[string][]string       `yaml:"release_def,omitempty" json:"release_def,omitempty"`
	ReleaseInfo *MetaReleaseInfo          `yaml:"release_info,omitempty" json:"release_info,omitempty"`
}

type MetaKitSetting struct {
	Stability map[string]string `yaml:"stability,omitempty" json:"stability,omitempty"`
	Type      string            `yaml:"type,omitempty" json:"type,omitempty"`
}

// Used by kit-info.json and version.json
type MetaReleaseInfo struct {
	Required []map[string]string `yaml:"required,omitempty" json:"required,omitempty"`
	Version  int                 `yaml:"version,omitempty" json:"version,omitempty"`
}

// kit-sha1.json structure
type MetaKitSha1 struct {
	Kits map[string]map[string]interface{} `yaml:"-,inline" json:"-,inline"`
}

// Instead of directly define the sha1 of a branch
// is possible define the depth and the sha1 together.
// This struct is used as value of the MetaKitSha1 map interface.
type MetaKitShaValue struct {
	Sha1  string `yaml:"sha1,omitempty" json:"sha1,omitempty"`
	Depth *int   `yaml:"depth,omitempty" json:"depth,omitempty"`
}

type KitReleaseSpec struct {
	Release *KitRelease `yaml:"release,omitempty" json:"release,omitempty"`

	File string `yaml:"-" json:"-"`
}

type KitRelease struct {
	Sources []*ReposcanKit   `yaml:"sources,omitempty" json:"sources,omitempty"`
	Target  KitReleaseTarget `yaml:"target,omitempty" json:"target,omitempty"`
	MainKit string           `yaml:"main_kit,omitempty" json:"main_kit,omitempty"`
}

type KitReleaseTarget struct {
	Name   string `yaml:"name" json:"name"`
	Url    string `yaml:"url,omitempty" json:"url,omitempty"`
	Branch string `yaml:"branch" json:"branch"`
}
