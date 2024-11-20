/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

type ReposcanAnalysis struct {
	Kits []*ReposcanKit `yaml:"kits,omitempty" json:"kits,omitempty"`
}

type ReposcanKit struct {
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	Url        string `yaml:"url,omitempty" json:"url,omitempty"`
	Branch     string `yaml:"branch,omitempty" json:"branch,omitempty"`
	CommitSha1 string `yaml:"commit_sha1,omitempty" json:"commit_sha1,omitempty"`
}
