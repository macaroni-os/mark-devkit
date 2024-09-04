/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernelspecs

import (
	"encoding/json"
	"strings"
)

func NewInitrdImage() *InitrdImage {
	return &InitrdImage{}
}

func NewInitrdImageFromFile(t *KernelType, file string) (*InitrdImage, error) {
	ans := NewInitrdImage()
	ans.Filename = file

	iprefix := t.InitrdPrefix
	if t.InitrdPrefix == "" {
		iprefix = "initramfs"
	}

	ans.Prefix = iprefix

	// Skip prefix + '-'
	file = file[len(iprefix)+1:]

	if t.Type != "" {
		ans.KernelType = t.Type
		file = file[len(t.Type)+1:]
	}

	words := strings.Split(file, "-")
	i := 0
	if t.WithArch {
		file = file[len(words[i])+1:]
		ans.Arch = words[i]
		i += 1
	}

	if len(words) > 3 {
		// POST: the version could contains a suffix
		ans.Version = words[i] + "-" + words[i+1]
		file = file[len(words[i])+1+len(words[i+1]):]
		i += 1
	} else {
		ans.Version = words[i]
		file = file[len(words[i]):]
	}

	if t.Suffix != "" && file != "" {
		ans.Suffix = file[1:]
	}

	return ans, nil
}

func (i *InitrdImage) SetPrefix(p string)     { i.Prefix = p }
func (i *InitrdImage) SetSuffix(s string)     { i.Suffix = s }
func (i *InitrdImage) SetKernelType(t string) { i.KernelType = t }
func (i *InitrdImage) SetArch(a string)       { i.Arch = a }
func (i *InitrdImage) SetVersion(v string)    { i.Version = v }
func (i *InitrdImage) SetFilename(f string)   { i.Filename = f }

func (i *InitrdImage) GetPrefix() string     { return i.Prefix }
func (i *InitrdImage) GetSuffix() string     { return i.Suffix }
func (i *InitrdImage) GetKernelType() string { return i.KernelType }
func (i *InitrdImage) GetArch() string       { return i.Arch }
func (i *InitrdImage) GetVersion() string    { return i.Version }
func (i *InitrdImage) GetFilename() string   { return i.Filename }

func (i *InitrdImage) String() string {
	data, _ := json.Marshal(i)
	return string(data)
}

func (i *InitrdImage) EqualTo(in *InitrdImage) bool {
	if i.Prefix != in.Prefix {
		return false
	}

	if i.KernelType != in.KernelType {
		return false
	}

	if i.Arch != in.Arch {
		return false
	}

	if i.Version != in.Version {
		return false
	}

	return true
}

func (i *InitrdImage) GenerateFilename() string {

	iprefix := i.Prefix
	if i.Prefix == "" {
		iprefix = "initramfs"
	}

	ans := iprefix + "-"

	if i.KernelType != "" {
		ans += i.KernelType
	}

	if i.Arch != "" {
		ans += "-" + i.Arch
	}

	if i.Version != "" {
		ans += "-" + i.Version
	}

	if i.Suffix != "" {
		ans += "-" + i.Suffix
	}

	return ans
}
