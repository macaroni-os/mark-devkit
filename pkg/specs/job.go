/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"

	"github.com/macaroni-os/macaronictl/pkg/utils"
	"gopkg.in/yaml.v3"
)

func (j *JobRendered) Yaml() ([]byte, error) {
	return yaml.Marshal(j)
}

func (j *JobRendered) GetPreChrootHooks() *[]*Hook {
	ans := []*Hook{}
	for _, hf := range j.HookFile {
		for idx, h := range hf.Hooks {
			if h.Type == HookOuterPreChroot {
				ans = append(ans, &hf.Hooks[idx])
			}
		}
	}

	return &ans
}

func (j *JobRendered) GetPostChrootHooks() *[]*Hook {
	ans := []*Hook{}
	for _, hf := range j.HookFile {
		for idx, h := range hf.Hooks {
			if h.Type == HookOuterPostChroot {
				ans = append(ans, &hf.Hooks[idx])
			}
		}
	}

	return &ans
}

func (j *JobRendered) GetOptionsEnvsMap() map[string]string {
	ans := make(map[string]string, 0)

	for k, v := range j.Options {
		// Bash doesn't support variable with dash.
		// I will convert dash with underscore.
		if strings.Contains(k, "-") {
			k = strings.ReplaceAll(k, "-", "_")
		}

		switch v.(type) {
		case int:
			ans[k] = fmt.Sprintf("%s", v.(int))
		case string:
			ans[k] = v.(string)
		default:
			continue
		}
	}

	for _, ev := range j.Envs {
		ans[ev.Key] = ev.Value
	}

	return ans
}

func (j *JobRendered) GetBindsMap() map[string]string {
	ans := make(map[string]string, 0)

	for _, b := range j.ChrootBinds {
		// TODO manage relative paths
		ans[b.Source] = b.Target
	}

	return ans
}

func (j *Job) Render(specfile string) (*JobRendered, error) {
	var err error
	ans := &JobRendered{
		Job:      j,
		HookFile: []*HookFile{},
	}

	specfileAbspath, err := filepath.Abs(specfile)
	if err != nil {
		return nil, err
	}

	// Render source.uri if is valid
	if j.Source.Uri != "" {
		ans.Source.Uri, err = helpers.RenderContentWithTemplates(
			j.Source.Uri,
			"", "", "source.uri", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render source.path if is valid
	if j.Source.Path != "" {
		ans.Source.Path, err = helpers.RenderContentWithTemplates(
			j.Source.Path,
			"", "", "source.path", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render source.target if is valid
	if j.Source.Target != "" {
		ans.Source.Target, err = helpers.RenderContentWithTemplates(
			j.Source.Target,
			"", "", "source.target", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
		// Resolve target dir as abs path if it's a relative path
		if !strings.HasPrefix(ans.Source.Target, "/") {
			ans.Source.Target = filepath.Join(
				filepath.Dir(specfileAbspath),
				ans.Source.Target,
			)
		}
	}

	// Render output.name
	if j.Output.Name != "" {
		ans.Output.Name, err = helpers.RenderContentWithTemplates(
			j.Output.Name,
			"", "", "output.name", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}
	}

	// Render output.dir
	if j.Output.Dir != "" {
		ans.Output.Dir, err = helpers.RenderContentWithTemplates(
			j.Output.Dir,
			"", "", "output.dir", j.Options, []string{},
		)
		if err != nil {
			return nil, err
		}

		// Resolve target dir as abs path if it's a relative path
		if !strings.HasPrefix(ans.Output.Dir, "/") {
			ans.Output.Dir = filepath.Join(
				filepath.Dir(specfileAbspath),
				ans.Output.Dir,
			)
		}
	}

	// Ensure workspacedir is an abs path
	if !strings.HasPrefix(j.WorkspaceDir, "/") {
		ans.WorkspaceDir = filepath.Join(
			filepath.Dir(specfileAbspath),
			j.WorkspaceDir,
		)
	}

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
