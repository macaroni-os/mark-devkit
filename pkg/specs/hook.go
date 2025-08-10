/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"gopkg.in/yaml.v3"
)

func NewHookFromYaml(data []byte, file string) (*Hook, error) {
	ans := &Hook{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	if ans.Type == "" {
		ans.Type = "inner-chroot"
	}

	return ans, nil
}

func (h *Hook) GetType() HookType       { return h.Type }
func (h *Hook) GetName() string         { return h.Name }
func (h *Hook) GetDescription() string  { return h.Description }
func (h *Hook) GetFile() string         { return h.File }
func (h *Hook) GetCommands() []string   { return h.Commands }
func (h *Hook) GetEntrypoint() []string { return h.Entrypoint }
func (h *Hook) GetBinds() []Bind        { return h.Binds }
