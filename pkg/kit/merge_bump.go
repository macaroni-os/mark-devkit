/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (m *MergeBot) BumpAtoms(mkit *specs.MergeKit, opts *MergeBotOpts) error {

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

	// We need to retrieve the head ref
	// before create branches to avoid
	// propagation of commits of different
	// branches
	headRef, err := repo.Head()
	if err != nil {
		return err
	}

	for pkg, files := range m.files4Commit {

		if opts.PullRequest {

			prBranchName := GetPrBranchNameForPkgBump(pkg, kit.Branch)
			prBranchExists, err := BranchExists(kit.Url, prBranchName)
			if err != nil {
				return err
			}

			if prBranchExists {
				// PR is already been pushed.
				m.Logger.InfoC(fmt.Sprintf(
					"[%s] PR branch already present. Nothing to do.",
					pkg))
				m.branches2Skip[prBranchName] = true

				// Restore committed files in order to avoid
				// that the same changes will be added in new commit.
				err = m.restoreFiles(kitDir, files, opts, worktree)
				if err != nil {
					return err
				}
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

			m.Logger.Info(fmt.Sprintf(":factory:[%s] Created branch %s.",
				pkg, prBranchName))

		}

		cMsg := fmt.Sprintf("Bump %s", pkg)
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

func (m *MergeBot) restoreFiles(kitDir string, files []string,
	opts *MergeBotOpts, worktree *git.Worktree) error {

	filesBase := []string{}

	for _, file := range files {
		// Drop kitDir prefix
		filesBase = append(filesBase, file[len(kitDir)+1:len(file)])
	}

	return worktree.Reset(&git.ResetOptions{
		Mode:  git.HardReset,
		Files: filesBase,
	})
}

func (m *MergeBot) commitFiles(kitDir string, files []string,
	commitMessage string, opts *MergeBotOpts,
	worktree *git.Worktree) (plumbing.Hash, error) {

	for _, file := range files {
		// Drop kitDir prefix
		f := file[len(kitDir)+1 : len(file)]
		_, err := worktree.Add(f)
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf(
				"error on commit file %s (%s): %s",
				file, f, err.Error())
		}
	}

	gOpts := &CloneOptions{
		SignatureName:  opts.SignatureName,
		SignatureEmail: opts.SignatureEmail,
	}

	return worktree.Commit(commitMessage,
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  gOpts.GetSignatureName(),
				Email: gOpts.GetSignatureEmail(),
				When:  time.Now(),
			},
		},
	)
}
