/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"gopkg.in/yaml.v3"
)

func NewMetroSpec() *MetroSpec {
	return &MetroSpec{
		Version: "1.0",
		Jobs:    []Job{},
	}
}

func (s *MetroSpec) LoadYaml(data []byte, file string) error {
	if err := yaml.Unmarshal(data, s); err != nil {
		return err
	}
	s.File = file
	return nil
}

func (s *MetroSpec) GetJob(name string) *Job {
	for idx := range s.Jobs {
		if s.Jobs[idx].Name == name {
			return &s.Jobs[idx]
		}
	}
	return nil
}
