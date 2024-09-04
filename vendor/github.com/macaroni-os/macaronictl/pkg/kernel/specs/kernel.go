/*
Copyright © 2021 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernelspecs

import (
	"encoding/json"
	"strings"
)

func NewKernelImage() *KernelImage {
	return &KernelImage{}
}

func NewKernelImageFromFile(t *KernelType, file string) (*KernelImage, error) {
	ans := NewKernelImage()
	ans.Filename = file

	kprefix := t.KernelPrefix
	if t.KernelPrefix == "" {
		kprefix = "kernel"
	}

	ans.Prefix = kprefix

	// Skip prefix + '-'
	file = file[len(kprefix)+1:]

	if t.Type != "" {
		ans.Type = t.Type
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

	if t.Suffix != "" {
		ans.Suffix = file[1:]
	}

	return ans, nil
}

func (k *KernelImage) SetPrefix(p string)   { k.Prefix = p }
func (k *KernelImage) SetSuffix(s string)   { k.Suffix = s }
func (k *KernelImage) SetVersion(v string)  { k.Version = v }
func (k *KernelImage) SetArch(v string)     { k.Arch = v }
func (k *KernelImage) SetType(t string)     { k.Type = t }
func (k *KernelImage) SetFilename(f string) { k.Filename = f }

func (k *KernelImage) GetPrefix() string   { return k.Prefix }
func (k *KernelImage) GetSuffix() string   { return k.Suffix }
func (k *KernelImage) GetVersion() string  { return k.Version }
func (k *KernelImage) GetArch() string     { return k.Arch }
func (k *KernelImage) GetType() string     { return k.Type }
func (k *KernelImage) GetFilename() string { return k.Filename }

func (k *KernelImage) String() string {
	data, _ := json.Marshal(k)
	return string(data)
}

func (k *KernelImage) EqualTo(i *KernelImage) bool {
	if k.Type != i.Type {
		return false
	}

	if k.Prefix != i.Prefix {
		return false
	}

	if k.Suffix != i.Suffix {
		return false
	}

	if k.Version != i.Version {
		return false
	}

	if k.Arch != i.Arch {
		return false
	}

	return true
}

func (k *KernelImage) GenerateFilename() string {

	kprefix := k.Prefix
	if k.Prefix == "" {
		kprefix = "kernel"
	}

	ans := kprefix + "-"

	if k.Type != "" {
		ans += k.Type
	}

	if k.Arch != "" {
		ans += "-" + k.Arch
	}

	if k.Version != "" {
		ans += "-" + k.Version
	}

	if k.Suffix != "" {
		ans += "-" + k.Suffix
	}

	return ans
}
