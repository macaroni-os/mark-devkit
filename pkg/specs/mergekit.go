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

func NewMergeKit() *MergeKit {
	return &MergeKit{
		Sources: []*ReposcanKit{},
	}
}

func (m *MergeKit) LoadFile(file string) error {
	// Read specfile
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return m.LoadYaml(content, file)
}

func (m *MergeKit) LoadYaml(data []byte, file string) error {
	if err := yaml.Unmarshal(data, m); err != nil {
		return err
	}
	m.File = file
	return nil
}

func (m *MergeKit) GetTargetKit() (*ReposcanKit, error) {
	ans := &ReposcanKit{
		Name:   m.Target.Name,
		Url:    m.Target.Url,
		Branch: m.Target.Branch,
	}

	sourceKit := m.GetSourceKit(m.Target.Name)

	if m.Target.Url == "" {
		if sourceKit == nil {
			return nil, fmt.Errorf("Target repo url not present!")
		}

		m.Target.Url = sourceKit.Url
	}

	return ans, nil
}

func (m *MergeKit) GetSourceKit(name string) *ReposcanKit {
	if len(m.Sources) > 0 {
		for idx := range m.Sources {
			if m.Sources[idx].Name == name {
				return m.Sources[idx]
			}
		}
	}

	return nil
}

func (m *MergeKit) GetEclassesInclude() *map[string][]string {
	var ans *map[string][]string = nil

	if m.Target.Eclasses != nil {
		ans = &m.Target.Eclasses.Include
	}
	return ans
}
func (m *MergeKit) GetFixupsInclude() *[]*MergeKitFixupInclude {
	var ans *[]*MergeKitFixupInclude = nil

	if m.Target.Fixups != nil {
		ans = &m.Target.Fixups.Include
	}

	return ans
}

func (m *MergeKit) GetMetadata() *MergeKitMetadata {
	if m.Target.Metadata == nil {
		m.Target.Metadata = &MergeKitMetadata{
			LayoutMasters: "",
		}
	}
	return m.Target.Metadata
}

func (m *MergeKitMetadata) GetLayoutMasters() string {
	if m.LayoutMasters == "" {
		return "core-kit"
	}
	return m.LayoutMasters
}
