/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"path/filepath"
	"strings"
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

	for pkg, files := range m.files4Commit {

		if opts.PullRequest {

			prBranchName := fmt.Sprintf(
				"%s%s-%s",
				prBranchPrefix, "bump",
				strings.ReplaceAll(strings.ReplaceAll(pkg, ".", "_"),
					"/", "_"),
			)

			prBranchExists, err := BranchExists(kit.Url, prBranchName)
			if err != nil {
				return err
			}

			if prBranchExists {
				// PR is already been pushed.
				m.Logger.InfoC(fmt.Sprintf(
					"[%s] PR branch already present. Nothing to do.",
					pkg))
				continue
			}

			headRef, err := repo.Head()
			if err != nil {
				return err
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
				Force:  true,
			}
			err := worktree.Checkout(&branchCoOpts)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
