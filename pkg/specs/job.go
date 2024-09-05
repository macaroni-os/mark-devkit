/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
	"path/filepath"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"

	"github.com/macaroni-os/macaronictl/pkg/utils"
	"gopkg.in/yaml.v3"
)

func (j *JobRendered) Yaml() ([]byte, error) {
	return yaml.Marshal(j)
}

func (j *Job) Render(specfile string) (*JobRendered, error) {
	var err error
	ans := &JobRendered{
		Job:      j,
		HookFile: []*HookFile{},
	}

	// Render source.uri if is valid
	if j.Source.Uri != "" {
		j.Source.Uri, err = helpers.RenderContentWithTemplates(
			j.Source.Uri,
			"", "", "source.uri", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render source.path if is valid
	if j.Source.Path != "" {
		j.Source.Path, err = helpers.RenderContentWithTemplates(
			j.Source.Path,
			"", "", "source.path", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render output.name
	if j.Output.Name != "" {
		j.Output.Name, err = helpers.RenderContentWithTemplates(
			j.Output.Name,
			"", "", "output.name", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render output.dir
	if j.Output.Dir != "" {
		j.Output.Dir, err = helpers.RenderContentWithTemplates(
			j.Output.Dir,
			"", "", "output.dir", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	specfileAbspath, err := filepath.Abs(specfile)
	hooksbasedir := filepath.Join(
		filepath.Dir(specfileAbspath),
		j.HooksBasedir,
	)

	for _, file := range j.Hooks {

		hf := filepath.Join(hooksbasedir, file)
		if !utils.Exists(hf) {
			return nil, fmt.Errorf("File %s not present", hf)
		}

		hookfile, err := NewHookFileFromFile(hf, j.Options)
		if err != nil {
			return nil, err
		}

		ans.HookFile = append(ans.HookFile, hookfile)
	}

	return ans, nil
}
