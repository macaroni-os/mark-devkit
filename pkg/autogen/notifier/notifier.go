/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package notifier

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

type Notifier interface {
	Notify(atom *specs.AutogenAtom, msg string) error
	GetType() string
	GetHook() *specs.MarkDevkitHook
}

func NewNotifier(t string, opts map[string]string, h *specs.MarkDevkitHook) (Notifier, error) {
	switch t {
	case specs.NotifyDiscord:
		return NewNotifierDiscord(opts, h), nil
	default:
		return nil, fmt.Errorf("Invalid notify type %s", t)
	}
}
