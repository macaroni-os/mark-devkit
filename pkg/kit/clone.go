/*
	Copyright © 2024 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package kit

import (
	"fmt"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type CloneOptions struct {
	GitCloneOptions *git.CloneOptions
	Verbose         bool
	Summary         bool
	Results         []*specs.ReposcanKit
}

func Clone(k *specs.ReposcanKit, targetdir string, o *CloneOptions) error {
	var r *git.Repository
	var err error
	log := logger.GetDefaultLogger()

	opts := *o.GitCloneOptions

	opts.URL = k.Url
	if k.Branch != "" {
		branchRefName := plumbing.NewBranchReferenceName(k.Branch)
		opts.ReferenceName = plumbing.ReferenceName(branchRefName)
	}

	if utils.Exists(targetdir) {
		// POST: The directory exists. I try to pull updates.

		pOpts := &git.PullOptions{
			SingleBranch:  opts.SingleBranch,
			Progress:      opts.Progress,
			Force:         true,
			RemoteName:    opts.RemoteName,
			ReferenceName: opts.ReferenceName,
			Depth:         opts.Depth,
		}

		r, err = git.PlainOpen(targetdir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Pull(pOpts)
		if err == git.NoErrAlreadyUpToDate {
			log.Info(fmt.Sprintf(":check_mark_button:[%s] already up to date.",
				k.Name))
		} else if err != nil {
			return fmt.Errorf("error on pull repo %s for branch %s: %s",
				k.Name, k.Branch, err.Error())
		}
	} else {

		r, err = git.PlainClone(targetdir, false, &opts)
		if err != nil {
			return err
		}

	}

	if k.CommitSha1 != "" {
		w, err := r.Worktree()
		if err != nil {
			return err
		}

		chOpts := &git.CheckoutOptions{
			Hash:  plumbing.NewHash(k.CommitSha1),
			Force: true,
		}

		err = w.Checkout(chOpts)
		if err != nil {
			return fmt.Errorf("error on checkout repo %s for branch %s and hash %s: %s",
				k.Name, k.Branch, k.CommitSha1, err.Error())
		}
	}

	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf(":right_arrow: [%s] @ %s",
		k.Name, ref.Hash()))
	if o.Verbose {
		commit, err := r.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		log.InfoC(fmt.Sprintf("%s", commit))
	}

	if o.Summary {
		res := *k
		res.CommitSha1 = fmt.Sprintf("%s", ref.Hash())
		o.Results = append(o.Results, &res)
	}

	return nil
}
