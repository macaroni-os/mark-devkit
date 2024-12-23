/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func NewKitReleaseSpec() *KitReleaseSpec {
	return &KitReleaseSpec{
		Release: &KitRelease{
			Sources: []*ReposcanKit{},
		},
	}
}

func (r *KitReleaseSpec) LoadFile(file string) error {
	// Read specfile
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return r.LoadYaml(content, file)
}

func (r *KitReleaseSpec) LoadYaml(data []byte, file string) error {
	if err := yaml.Unmarshal(data, r); err != nil {
		return err
	}
	r.File = file
	return nil
}

func (r *KitReleaseSpec) GetSourceKit(name string) *ReposcanKit {
	if len(r.Release.Sources) > 0 {
		for idx := range r.Release.Sources {
			if r.Release.Sources[idx].Name == name {
				return r.Release.Sources[idx]
			}
		}
	}

	return nil
}

func (r *KitRelease) GetTargetKit() (*ReposcanKit, error) {
	ans := &ReposcanKit{
		Name:   r.Target.Name,
		Url:    r.Target.Url,
		Branch: r.Target.Branch,
	}

	if r.Target.Url == "" {
		return nil, fmt.Errorf("Target repo url not present!")
	}

	return ans, nil
}

func (r *KitRelease) GetMainKit() string {
	if r.MainKit == "" {
		return "core-kit"
	}
	return r.MainKit
}

func (r *MetaReleaseInfo) Json() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (i *MetaKitInfo) Json() ([]byte, error) {
	return json.MarshalIndent(i, "", "  ")
}

func (i *MetaKitSha1) Json() ([]byte, error) {
	return json.MarshalIndent(i.Kits, "", " ")
}
