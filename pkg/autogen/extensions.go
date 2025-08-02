/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/autogen/extensions"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

func (a *AutogenBot) ConsumeExtensions(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom,
	autogenDef *specs.AutogenDefinition,
	mapref *map[string]interface{}) error {

	for _, atomExt := range atom.Extensions {

		// Retrieve extension options
		extOpts, err := autogenDef.GetExtensionOptions(atomExt)
		if err != nil {
			a.Logger.Error("[%s] %s", atom.Name, err.Error())
			return err
		}

		extOpts.Options["download_dir"] = a.GetDownloadDir()
		extOpts.Options["workdir"] = a.WorkDir
		extOpts.Options["specfile"] = aspec.File

		a.Logger.Info(
			fmt.Sprintf(":brain:[%s] Elaborating extension %s...",
				atom.Name, atomExt))

		err = a.ConsumeExtension(mkit, aspec, atom, def,
			mapref, extOpts,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AutogenBot) ConsumeExtension(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom,
	mapref *map[string]interface{},
	extensionDef *specs.AutogenExtension) error {

	ext, err := extensions.NewExtension(extensionDef.Name, extensionDef.Options)
	if err != nil {
		return err
	}

	// Execute extension code
	err = ext.Elaborate(atom, def, mapref)

	return err
}
