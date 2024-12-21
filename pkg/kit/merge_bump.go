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

	for pkg, files := range m.files4Commit {
		cMsg := fmt.Sprintf("Bump %s", pkg)
		commitHash, err := m.commitFiles(kitDir, files, cMsg, opts, worktree)
		if err != nil {
			return err
		}

		if opts.Verbose {
			commit, _ := repo.CommitObject(commitHash)
			m.Logger.InfoC(fmt.Sprintf("%s", commit))
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
