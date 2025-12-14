/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/autogen/notifier"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
)

func (a *AutogenBot) setupNotifiers() error {

	opts := make(map[string]string, 0)
	opts[guard_specs.ServiceRateLimiter] = "1"

	for _, hook := range a.Config.Notifier.Hooks {
		if hook.Enable {

			a.Logger.Debug(fmt.Sprintf(
				":factory: Initializing notifier %s..", hook.Name))
			n, err := notifier.NewNotifier(hook.Type, opts, hook)
			if err != nil {
				return err
			}
			a.Notifiers = append(a.Notifiers, n)
		}
	}

	return nil
}

func (a *AutogenBot) Notify(atom *specs.AutogenAtom, msg string) error {
	var ans error = nil

	for _, hook := range a.Notifiers {
		err := hook.Notify(atom, msg)
		if err != nil && ans == nil {
			ans = err
		}

		if err != nil {
			a.Logger.Warning(fmt.Sprintf(
				":fire: Notifier %s: %s", hook.GetHook().Name, err.Error()))
		}
	}

	return ans
}
