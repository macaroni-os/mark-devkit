/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

type HookType string

const (
	HookInnerChroot     = "inner-chroot"
	HookOuterChroot     = "outer-chroot"
	HookOuterPostChroot = "outer-post-chroot"
	HookOuterPreChroot  = "outer-pre-chroot"
)

type MetroSpec struct {
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Jobs    []Job  `yaml:"jobs,omitempty" json:"jobs,omitempty"`

	File string `yaml:"-" json:"-"`
}

type Job struct {
	Name   string    `yaml:"name" json:"name"`
	Source JobSource `yaml:"source,omitempty" json:"source,omitempty"`
	Output JobOutput `yaml:"output,omitempty" json:"omitempty,omitempty"`

	Options map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
	Envs    []*EnvVar              `yaml:"environments,omitempty" json:"environments,omitempty"`

	ChrootBinds []Bind `yaml:"chroot_binds,omitempty" json:"chroot_binds,omitempty"`

	WorkspaceDir string   `yaml:"workspacedir,omitempty" json:"workspacedir,omitempty"`
	HooksBasedir string   `yaml:"hooks_basedir,omitempty" json:"hooks_basedir,omitempty"`
	Hooks        []string `yaml:"hooks_files,omitempty" json:"hooks_files,omitempty"`
}

type JobRendered struct {
	*Job
	HookFile []*HookFile `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type JobSource struct {
	Type   string `yaml:"type,omitempty" json:"type,omitempty"`
	Uri    string `yaml:"uri,omitempty" json:"uri,omitempty"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
}

type JobOutput struct {
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Dir  string `yaml:"dir,omitempty" json:"dir,omitempty"`
}

type Bind struct {
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
}

type Hook struct {
	File        string   `yaml:"-" json:"-"`
	Name        string   `yaml:"name,omitempty" json:"name,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Type        HookType `yaml:"type,omitempty" json:"type,omitempty"`
	Commands    []string `yaml:"commands,omitempty" json:"commands,omitempty"`
	Entrypoint  []string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`

	Binds []Bind `yaml:"chroot_binds,omitempty" json:"chroot_binds,omitempty"`
}

type HookFile struct {
	File  string `yaml:"-" json:"-"`
	Hooks []Hook `yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

type EnvVar struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}
