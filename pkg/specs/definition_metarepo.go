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
	Version  int                 `yaml:version,omitempty" json:"version,omitempty"`
}

// kit-sha1.json structure
type MetaKitSha1 struct {
	Kits map[string]map[string]string `yaml:"-,inline" json:",inline"`
}
