/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
	"os"
	"strings"

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
			LayoutMasters:          "",
			Aliases:                []string{},
			ManifestHashes:         []string{},
			ManifestRequiredHashes: []string{},
		}
	}
	return m.Target.Metadata
}

func (m *MergeKitTarget) GetThirdpartyMirrorsUris(alias string) []string {
	ans := []string{}

	if len(m.ThirdpartyMirrors) > 0 {
		for idx := range m.ThirdpartyMirrors {
			if m.ThirdpartyMirrors[idx].Alias == alias {
				return m.ThirdpartyMirrors[idx].Uri
			}
		}
	}

	return ans
}

func (m *MergeKitMetadata) GetLayoutMasters() string {
	if m.LayoutMasters == "" {
		return "core-kit"
	}
	return m.LayoutMasters
}

func (m *MergeKitMetadata) HasAliases() bool {
	return len(m.Aliases) > 0
}

func (m *MergeKitMetadata) HasManifestHashes() bool {
	return len(m.ManifestHashes) > 0
}

func (m *MergeKitMetadata) HashManifestReqHashes() bool {
	return len(m.ManifestRequiredHashes) > 0
}

func (f *MergeKitFixupInclude) GetType() string {
	ans := "file"
	if f.Dir != "" {
		ans = "directory"
	}

	return ans
}

func (f *MergeKitFixupInclude) GetName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.To
}

func NewDistfilesSpec() *DistfilesSpec {
	return &DistfilesSpec{
		MergeKit:        NewMergeKit(),
		FallbackMirrors: []*MergeKitThirdPartyMirror{},
	}
}

func (d *DistfilesSpec) LoadFile(file string) error {
	// Read specfile
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return d.LoadYaml(content, file)
}

func (d *DistfilesSpec) LoadYaml(data []byte, file string) error {
	if err := yaml.Unmarshal(data, d); err != nil {
		return err
	}
	d.File = file
	return nil
}

func (l *MirrorLayoutMode) GetAtomPath(fileName, fileSha512, fileBlake2b string) (ans string) {
	if l.Type == "flat" {
		ans = "/" + fileName
		return
	}

	if l.Type == "content-hash" {
		ans = "/"
		subpaths := strings.Split(l.HashMode, ":")
		for idx := range subpaths {
			if l.Hash == "SHA512" {
				ans += fileSha512[idx*2:idx*2+2] + "/"
			} else {
				ans += fileBlake2b[idx*2:idx*2+2] + "/"
			}
		}

		if l.Hash == "SHA512" {
			ans += fileSha512
		} else {
			ans += fileBlake2b
		}
	}

	return
}
