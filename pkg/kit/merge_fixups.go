/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/macaroni-os/macaronictl/pkg/utils"
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

	headRef, err := repo.Head()
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
		includeType := include.GetType()
		name := include.GetName()

		if m.IsANewBranch {
			cMsg = fmt.Sprintf("Add %s %s", includeType, name)
		} else {
			cMsg = fmt.Sprintf("Update %s %s", includeType, name)
		}

		if opts.PullRequest {
			// NOTE: pull request for a new branch it doesn't make sense
			// Probably we need to add a check.
			prBranchName := fmt.Sprintf(
				"%s%s-%s",
				prBranchPrefix, "fixup-include-",
				strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"),
					"/", "_"),
			)

			prBranchExists, err := BranchExists(kit.Url, prBranchName)
			if err != nil {
				return err
			}

			if prBranchExists {
				// PR is already been pushed.
				m.Logger.InfoC(fmt.Sprintf(
					"[%s] PR branch already present for fixup %s. Nothing to do.",
					prBranchName, name))
				continue
			}

			branchRef := plumbing.NewBranchReferenceName(prBranchName)
			ref := plumbing.NewHashReference(branchRef, headRef.Hash())
			// The created reference is saved in the storage.
			err = repo.Storer.SetReference(ref)
			if err != nil {
				return err
			}

			// Creating the new branch for the PR.
			branchCoOpts := git.CheckoutOptions{
				Branch: plumbing.ReferenceName(branchRef),
				Create: false,
				Keep:   true,
			}

			if err := worktree.Checkout(&branchCoOpts); err != nil {
				return err
			}

			m.fixupBranches[name] = include
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

		if opts.PullRequest {
			// Return to working branch
			targetBranchRef := plumbing.NewBranchReferenceName(kit.Branch)
			branchCoOpts := git.CheckoutOptions{
				Branch: plumbing.ReferenceName(targetBranchRef),
				Create: false,
				Keep:   true,
			}
			err := worktree.Checkout(&branchCoOpts)
			if err != nil {
				return err
			}

			// Restore committed files in order to avoid
			// that the same changes will be added in new commit.
			err = m.restoreFiles(kitDir, files, opts, worktree)
			if err != nil {
				return err
			}
		}
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
