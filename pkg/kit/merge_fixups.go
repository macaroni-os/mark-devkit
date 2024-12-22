/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

func (m *MergeBot) MergeFixups(mkit *specs.MergeKit, opts *MergeBotOpts) error {
	fixupIncludes := mkit.GetFixupsInclude()
	if fixupIncludes == nil {
		return nil
	}

	kit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)

	// Open the repository
	repo, err := git.PlainOpen(kitDir)
	if err != nil {
		return err
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	for _, include := range *fixupIncludes {
		files, err := m.mergeFixupInclude(kitDir, include, mkit, opts)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			continue
		}

		cMsg := ""
		includeType := "file"
		name := include.Name
		if include.Dir != "" {
			includeType = "directory"
		}

		if include.Name == "" {
			name = include.To
		}

		if m.IsANewBranch {
			cMsg = fmt.Sprintf("Add %s %s", includeType, name)
		} else {
			cMsg = fmt.Sprintf("Update %s %s", includeType, name)
		}

		commitHash, err := m.commitFiles(kitDir, files, cMsg, opts, worktree)
		if err != nil {
			return err
		}

		if opts.Verbose {
			commit, _ := repo.CommitObject(commitHash)
			m.Logger.InfoC(fmt.Sprintf("%s", commit))
		}

		m.hasCommit = true
	}

	return nil
}

func (m *MergeBot) mergeFixupInclude(targetKitDir string,
	include *specs.MergeKitFixupInclude,
	mkit *specs.MergeKit, opts *MergeBotOpts) ([]string, error) {

	ans := []string{}

	m.Logger.Debug(fmt.Sprintf("Checking fixup %s/%s for %s...",
		include.File, include.Dir, targetKitDir))

	specFileBasedir := filepath.Dir(mkit.File)

	if include.File != "" {

		sourceFile := filepath.Join(specFileBasedir, include.File)
		targetFile := filepath.Join(targetKitDir, include.To)
		targetFileMd5 := ""

		targetDir := filepath.Dir(targetFile)

		if !utils.Exists(targetDir) {
			err := helpers.EnsureDirWithoutIds(targetDir, 0755)
			if err != nil {
				return ans, err
			}
		}

		if !utils.Exists(sourceFile) {
			m.Logger.Warning(fmt.Sprintf(":warning:Fixup file %s not found. Skipped.",
				sourceFile))
			return ans, nil
		}

		sourceFileMd5, err := helpers.GetFileMd5(sourceFile)
		if err != nil {
			return ans, err
		}

		if utils.Exists(targetFile) {
			targetFileMd5, err = helpers.GetFileMd5(targetFile)
			if err != nil {
				return ans, err
			}
		}

		if sourceFileMd5 != targetFileMd5 {
			err := helpers.CopyFile(sourceFile, targetFile)
			if err != nil {
				return ans, err
			}

			ans = append(ans, targetFile)
		}

	} else {

		sourceDir := filepath.Join(specFileBasedir, include.Dir)
		targetDir := filepath.Join(targetKitDir, include.To)

		if !utils.Exists(sourceDir) {
			m.Logger.Warning(fmt.Sprintf(":warning:Fixup directory %s not found. Skipped.",
				sourceDir))
			return ans, nil
		}

		files, err := m.copyFilesDir(sourceDir, targetDir)
		if err != nil {
			return ans, err
		}

		ans = append(ans, files...)
	}

	return ans, nil
}
