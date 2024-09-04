/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernelspecs

import (
	"regexp"
)

type KernelImage struct {
	Filename string `json:"filename,omitempty" yaml:"filename,omitempty"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Prefix   string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Suffix   string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	Version  string `json:"version,omitempty" yaml:"version,omitempty"`
	Arch     string `json:"arch,omitempty" yaml:"arch,omitempty"`
}

type InitrdImage struct {
	Filename   string `json:"filename,omitempty" yaml:"filename,omitempty"`
	Prefix     string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Suffix     string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	KernelType string `json:"kernel_type,omitempty" yaml:"kernel_type,omitempty"`
	Arch       string `json:"arch,omitempty" yaml:"arch,omitempty"`
	Version    string `json:"version,omitempty" yaml:"version,omitempty"`
}

type KernelType struct {
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	KernelPrefix string `json:"kernel_prefix,omitempty" yaml:"kernel_prefix,omitempty"`
	InitrdPrefix string `json:"initrd_prefix,omitempty" yaml:"initrd_prefix,omitempty"`
	Suffix       string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	WithArch     bool   `json:"with_arch,omitempty" yaml:"with_arch,omitempty"`

	Regex *regexp.Regexp `json:"-" yaml:"-"`
}

type KernelFiles struct {
	Kernel *KernelImage `json:"kernel,omitempty" yaml:"kernel,omitempty"`
	Initrd *InitrdImage `json:"initrd,omitempty" yaml:"initrd,omitempty"`
	Type   *KernelType  `json:"type,omitempty" yaml:"type,omitempty"`
}

type BootFiles struct {
	Dir   string         `json:"dir,omitempty" yaml:"dir,omitempty"`
	Files []*KernelFiles `json:"files,omitempty" yaml:"files,omitempty"`

	BzImageLink string `json:"bzImage,omitempty" yaml:"bzImage,omitempty"`
	InitrdLink  string `json:"initrd,omitempty" yaml:"initrd,omitempty"`
}
