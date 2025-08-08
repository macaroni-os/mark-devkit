/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
)

type Extension interface {
	Elaborate(restGuard *guard.RestGuard,
		atom, def *specs.AutogenAtom,
		mapref *map[string]interface{}) error
	GetName() string
}

func NewExtension(t string, opts map[string]string) (Extension, error) {
	switch t {
	case specs.ExtensionCustom:
		return NewExtensionCustom(opts)
	case specs.ExtensionGolang:
		return NewExtensionGolang(opts)
	case specs.ExtensionRust:
		return NewExtensionRust(opts)
	case specs.ExtensionGitSubmodules:
		return NewExtensionGitSubmodules(opts)
	default:
		return nil, fmt.Errorf("Invalid extension %s", t)
	}
}
