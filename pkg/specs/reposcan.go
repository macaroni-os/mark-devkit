/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

func NewReposcanAnalysis(file string) (*ReposcanAnalysis, error) {
	ans := &ReposcanAnalysis{}

	content, err := os.ReadFile(file)
	if err != nil {
		return ans, err
	}

	err = yaml.Unmarshal(content, ans)
	return ans, err
}

func (ra *ReposcanAnalysis) Yaml() ([]byte, error) {
	return yaml.Marshal(ra)
}

func (ra *ReposcanAnalysis) Json() ([]byte, error) {
	return json.Marshal(ra)
}
