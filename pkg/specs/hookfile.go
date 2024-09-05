/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
)

func NewHookFileFromFile(file string, opts map[string]interface{}) (*HookFile, error) {
	// Read the main specification file
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	renderedHookfile, err := helpers.RenderContentWithTemplates(
		string(content),
		// TODO: Permit to define default/values file.
		"", "",
		file,
		opts,
		// TODO: add support for templates
		[]string{},
	)
	if err != nil {
		return nil, err
	}

	return NewHookFileFromYaml([]byte(renderedHookfile), file)
}

func NewHookFileFromYaml(data []byte, file string) (*HookFile, error) {
	ans := &HookFile{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	for idx := range ans.Hooks {
		ans.Hooks[idx].File = file
		if ans.Hooks[idx].Type == "" {
			ans.Hooks[idx].Type = "inner-chroot"
		}
	}

	return ans, nil
}

func (h *HookFile) GetFile() string   { return h.File }
func (h *HookFile) GetHooks() *[]Hook { return &h.Hooks }
