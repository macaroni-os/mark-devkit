/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

const (
	layoutConfTemplate = `repo_name = %s
thin-manifests = true
sign-manifests = false
profile-formats = portage-2
cache-formats = md5-dict
`
)

func (m *MergeBot) prepareMetadataDir(mkit *specs.MergeKit,
	opts *MergeBotOpts) error {
	var err error
	kit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)
	metadataDir := filepath.Join(kitDir, "metadata")

	if !utils.Exists(metadataDir) {
		err = helpers.EnsureDirWithoutIds(metadataDir, 0755)
		if err != nil {
			return err
		}
	}

	metadata := mkit.GetMetadata()

	layoutData := fmt.Sprintf(layoutConfTemplate, kit.Name)

	if metadata.GetLayoutMasters() != kit.Name {
		// core-kit doesn't need masters
		layoutData += fmt.Sprintf("masters = %s\n", metadata.GetLayoutMasters())
	}
	if metadata.HasAliases() {
		layoutData += fmt.Sprintf("aliases = %s\n",
			strings.Join(metadata.Aliases, " "))
	}
	if metadata.HasManifestHashes() {
		layoutData += fmt.Sprintf("manifest-hashes = %s\n",
			strings.Join(metadata.ManifestHashes, " "))
	}
	if metadata.HashManifestReqHashes() {
		layoutData += fmt.Sprintf("manifest-required-hashes = %s\n",
			strings.Join(metadata.ManifestRequiredHashes, " "))
	}

	layoutDataMd5 := fmt.Sprintf("%x", md5.Sum([]byte(layoutData)))

	layoutConf := filepath.Join(metadataDir, "layout.conf")
	layoutConfMd5 := ""
	cMsg := "Add metadata/layout.conf"

	if utils.Exists(layoutConf) {
		cMsg = "Update metadata/layout.conf"
		layoutConfMd5, err = helpers.GetFileMd5(layoutConf)
		if err != nil {
			return err
		}
	}

	if layoutDataMd5 != layoutConfMd5 {
		err = os.WriteFile(layoutConf, []byte(layoutData), 0644)
		if err != nil {
			return err
		}

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

		prBranchName := fmt.Sprintf(
			"%s%s",
			prBranchPrefix, "metadata-update",
		)

		// Restore committed files in order to avoid
		// that the same changes will be added in new commit.
		files := []string{layoutConf}
		defer m.restoreFiles(kitDir, files, opts, worktree)

		if opts.PullRequest {
			// NOTE: pull request for a new branch it doesn't make sense
			// Probably we need to add a check.
			prBranchExists, err := BranchExists(kit.Url, prBranchName)
			if err != nil {
				return err
			}

			if prBranchExists {
				// PR is already been pushed.
				m.Logger.InfoC(fmt.Sprintf(
					"[%s] PR branch already present for metadata. Nothing to do.",
					prBranchName))
				return nil
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

		commitHash, err := m.commitFiles(kitDir, files, cMsg, opts, worktree)
		if err != nil {
			return err
		}

		if opts.Verbose {
			commit, _ := repo.CommitObject(commitHash)
			m.Logger.InfoC(fmt.Sprintf("%s", commit))
		}

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

			m.metadataUpdate = true

		} else {
			m.hasCommit = true
		}
	}

	return nil
}
