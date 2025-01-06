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
	"sort"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (m *MergeBot) prepareProfilesDir(mkit *specs.MergeKit,
	candidates []*specs.RepoScanAtom,
	opts *MergeBotOpts) error {

	var err error
	kit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)
	profilesDir := filepath.Join(kitDir, "profiles")
	files4Commit := make(map[string]string, 0)

	// Retrieve the list of categories of the target resolver.
	categories := []string{}
	catMap := make(map[string]bool, 0)

	if !utils.Exists(profilesDir) {
		err = helpers.EnsureDirWithoutIds(profilesDir, 0755)
		if err != nil {
			return err
		}
	}

	if len(m.TargetResolver.Map) > 0 {
		for catpkg := range m.TargetResolver.Map {
			words := strings.Split(catpkg, "/")
			if len(words) != 2 {
				return fmt.Errorf("Invalid package name %s", catpkg)
			}
			catMap[words[0]] = true
		}

	}

	// Append candidites categories
	if len(candidates) > 0 {
		for _, atom := range candidates {
			catMap[atom.Category] = true
		}
	}

	if len(catMap) > 0 {
		for cat := range catMap {
			categories = append(categories, cat)
		}
	}

	// Sort categories to ensure always same format
	sort.Strings(categories)
	categoriesFileContent := strings.Join(categories, "\n") + "\n"

	categoriesFile := filepath.Join(profilesDir, "categories")
	categoriesDataMd5 := fmt.Sprintf("%x", md5.Sum([]byte(categoriesFileContent)))
	categoriesTargetMd5 := ""

	if utils.Exists(categoriesFile) {
		categoriesTargetMd5, err = helpers.GetFileMd5(categoriesFile)
		if err != nil {
			return err
		}
	}

	if categoriesDataMd5 != categoriesTargetMd5 {
		err = os.WriteFile(categoriesFile, []byte(categoriesFileContent), 0644)
		if err != nil {
			return err
		}
		if categoriesTargetMd5 == "" {
			files4Commit[categoriesFile] = "Add profiles/categories file"
		} else {
			files4Commit[categoriesFile] = "Update profiles/categories file"
		}
	}

	// Check repo_name file
	repoNamefile := filepath.Join(profilesDir, "repo_name")
	repoNameData := kit.Name + "\n"
	repoNameMd5 := fmt.Sprintf("%x", md5.Sum([]byte(repoNameData)))
	repoNameTargetMd5 := ""

	if utils.Exists(repoNamefile) {
		repoNameTargetMd5, err = helpers.GetFileMd5(repoNamefile)
		if err != nil {
			return err
		}
	}

	if repoNameMd5 != repoNameTargetMd5 {
		err = os.WriteFile(repoNamefile, []byte(repoNameData), 0644)
		if err != nil {
			return err
		}
		if repoNameTargetMd5 == "" {
			files4Commit[repoNamefile] = "Add profiles/repo_name file"
		} else {
			files4Commit[repoNamefile] = "Update profiles/repo_name file"
		}
	}

	if len(mkit.Target.ThirdpartyMirrors) > 0 {
		err = m.generateThirdPartyMirrorsFile(mkit, profilesDir, &files4Commit, opts)
		if err != nil {
			return err
		}
	}

	if len(files4Commit) > 0 {

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
			prBranchPrefix, "profiles-update",
		)

		// Restore committed files in order to avoid
		// that the same changes will be added in new commit.
		files := []string{}
		for f := range files4Commit {
			files = append(files, f)
		}
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
					"[%s] PR branch already present for profiles. Nothing to do.",
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

		for f, cMsg := range files4Commit {
			commitHash, err := m.commitFiles(kitDir, []string{f},
				cMsg, opts, worktree)
			if err != nil {
				return err
			}

			if opts.Verbose {
				commit, _ := repo.CommitObject(commitHash)
				m.Logger.InfoC(fmt.Sprintf("%s", commit))
			}
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

			m.profilesUpdate = true

		} else {
			m.hasCommit = true
		}
	}

	return nil
}

func (m *MergeBot) generateThirdPartyMirrorsFile(mkit *specs.MergeKit,
	profilesDir string, files4CommitMapRef *map[string]string,
	opts *MergeBotOpts) error {

	var err error
	targetFile := filepath.Join(profilesDir, "thirdpartymirrors")
	files4Commit := *files4CommitMapRef

	content := ""

	for idx := range mkit.Target.ThirdpartyMirrors {
		padding := "\t"
		if len(mkit.Target.ThirdpartyMirrors[idx].Alias) < 8 {
			padding += "\t"
		}

		content += fmt.Sprintf("%s%s%s\n",
			mkit.Target.ThirdpartyMirrors[idx].Alias,
			padding,
			strings.Join(mkit.Target.ThirdpartyMirrors[idx].Uri, " "))
	}

	contentMd5 := fmt.Sprintf("%x", md5.Sum([]byte(content)))
	targetFileMd5 := ""

	if utils.Exists(targetFile) {
		targetFileMd5, err = helpers.GetFileMd5(targetFile)
		if err != nil {
			return err
		}
	}

	if contentMd5 != targetFileMd5 {
		err = os.WriteFile(targetFile, []byte(content), 0644)
		if err != nil {
			return err
		}
		if targetFileMd5 == "" {
			files4Commit[targetFile] = "Add profiles/thirdpartymirrors file"
		} else {
			files4Commit[targetFile] = "Update profiles/thirdpartymirrors file"
		}
	}

	return nil
}
