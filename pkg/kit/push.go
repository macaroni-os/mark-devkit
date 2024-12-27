/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v68/github"
)

type PushOptions struct {
	Token      string
	RemoteName string
}

func NewPushOptions() *PushOptions {
	return &PushOptions{
		Token:      "",
		RemoteName: "origin",
	}
}

func getGithubAuth(opts *PushOptions) (*http.BasicAuth, error) {
	token := opts.Token
	if opts.Token == "" {
		// Retrieve token from env
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("Missing git token! Push interrupted!")
		}
	}

	auth := &http.BasicAuth{
		Username: "oauth2",
		Password: token,
	}

	return auth, nil
}

func Push(repoDir string, opts *PushOptions) error {
	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	auth, err := getGithubAuth(opts)
	if err != nil {
		return err
	}

	return repo.Push(&git.PushOptions{
		RemoteName: opts.RemoteName,
		Auth:       auth,
	})
}

func PushBranch(repoDir, branch string, opts *PushOptions) error {
	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	headRef, err := repo.Head()
	if err != nil {
		return err
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	currentBranch := headRef.Name().Short()

	// Checkout on specific branch
	if currentBranch != branch {
		targetBranchRef := plumbing.NewBranchReferenceName(branch)
		branchCoOpts := git.CheckoutOptions{
			Branch: plumbing.ReferenceName(targetBranchRef),
		}
		err := worktree.Checkout(&branchCoOpts)
		if err != nil {
			return err
		}
	}

	auth, err := getGithubAuth(opts)
	if err != nil {
		return err
	}
	err = repo.Push(&git.PushOptions{
		RemoteName: opts.RemoteName,
		Auth:       auth,
	})

	if currentBranch != branch {
		// Return to previous branch
		targetBranchRef := plumbing.NewBranchReferenceName(currentBranch)
		branchCoOpts := git.CheckoutOptions{
			Branch: plumbing.ReferenceName(targetBranchRef),
		}
		err := worktree.Checkout(&branchCoOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreatePullRequest(client *github.Client, ctx context.Context,
	title, srcBranch, targetBranch, body,
	githubUser, githubRepo string) (*github.PullRequest, error) {

	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(srcBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx,
		githubUser, githubRepo, newPR)
	return pr, err
}
