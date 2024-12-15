/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

func Push(repoDir string, opts *PushOptions) error {
	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	token := opts.Token
	if opts.Token == "" {
		// Retrieve token from env
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("Missing git token! Push interrupted!")
		}
	}

	auth := &http.BasicAuth{
		Username: "oauth2",
		Password: token,
	}

	return repo.Push(&git.PushOptions{
		RemoteName: opts.RemoteName,
		Auth:       auth,
	})
}
