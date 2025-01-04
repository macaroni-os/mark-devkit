/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

func (m *MergeBot) Clean(specfile string, opts *MergeBotOpts) error {
	// Load MergeKit specs
	mkit := specs.NewMergeKit()

	if opts.CleanWorkingDir {
		defer os.RemoveAll(m.WorkDir)
	}

	err := mkit.LoadFile(specfile)
	if err != nil {
		return err
	}

	targetKit, err := mkit.GetTargetKit()
	if err != nil {
		return err
	}

	m.Logger.InfoC(m.Logger.Aurora.Bold(
		fmt.Sprintf(":castle:Work directory:\t%s\n:rocket:Target Kit:\t\t%s",
			m.WorkDir, targetKit.Name)))

	if opts.PullSources {
		// Clone sources
		// NOTE: we need the sources directory
		//       in order to have all eclasses
		//       files needed for generate the
		//       reposcan file of the target kit.
		err = m.cloneSourcesKits(mkit, opts)
		if err != nil {
			return err
		}

		// Clone target kit
		err = m.cloneTargetKit(mkit, opts)
		if err != nil {
			return err
		}
	}

	// Generate kit-cache files
	if opts.GenReposcan {

		m.Logger.InfoC(m.Logger.Aurora.Bold(
			fmt.Sprintf(":brain:[%s] Generating reposcan file...",
				targetKit.Name)))

		err := helpers.EnsureDirWithoutIds(m.GetReposcanDir(), 0755)
		if err != nil {
			return err
		}

		// Prepare eclass dir list
		ra := &specs.ReposcanAnalysis{Kits: mkit.Sources}
		eclassDirs, err := ra.GetKitsEclassDirs(m.GetSourcesDir())
		if err != nil {
			return err
		}

		// Generate target
		sourceDir := filepath.Join(m.GetTargetDir(), targetKit.Name)
		targetFile := filepath.Join(m.GetReposcanDir(), "target-"+targetKit.Name+"-"+targetKit.Branch)
		err = m.GenerateKitCacheFile(sourceDir, targetKit.Name, targetKit.Branch,
			targetFile, eclassDirs, opts.Concurrency, true)
		if err != nil {
			return err
		}

	}

	// Setup target resolver
	err = m.SetupTargetResolver(mkit, opts, targetKit)
	if err != nil {
		return err
	}

	// Search candidates for remove
	candidates, err := m.SearchAtoms2Clean(mkit, opts)
	if err != nil {
		return err
	}

	if len(*candidates) > 0 {
		m.Logger.Info(fmt.Sprintf(":dart:Found %d candidates.",
			len(*candidates)))

		// Aggregate packages for PN in order to
		// avoid rebase issues with multiple commits.
		candidates4PN := make(map[string][]*specs.RepoScanAtom, 0)

		for _, candidate := range *candidates {
			m.Logger.Info(fmt.Sprintf(":pizza:[%s] %s",
				candidate.Kit, candidate.Atom))
			if pkgs, present := candidates4PN[candidate.CatPkg]; present {
				candidates4PN[candidate.CatPkg] = append(pkgs, candidate)
			} else {
				candidates4PN[candidate.CatPkg] = []*specs.RepoScanAtom{candidate}
			}
		}

		// Remove packages and update Manifest
		err = m.CleanAtoms(mkit, opts, &candidates4PN)
		if err != nil {
			return err
		}

		if opts.Push {
			// Push commits and/or PR
			err = m.PushRemoves(mkit, opts, &candidates4PN)
			if err != nil {
				return err
			}
		}

	} else {
		m.Logger.Info(
			":smiling_face_with_sunglasses:No candidates for remove found. Nothing to do.",
		)
	}

	return nil
}

func (m *MergeBot) CleanAtoms(mkit *specs.MergeKit, opts *MergeBotOpts,
	candidates4PN *map[string][]*specs.RepoScanAtom) error {

	for catpkg, candidates := range *candidates4PN {
		err := m.cleanAtom(mkit, opts, catpkg, candidates)
		if err != nil {
			// NOTE: for now i block process if something goes wrong.
			//       We can skip this package in the future probably.
			return err
		}
	}

	return nil
}

func (m *MergeBot) cleanAtom(mkit *specs.MergeKit, opts *MergeBotOpts,
	catpkg string,
	candidates []*specs.RepoScanAtom) error {

	var err error
	var manifest *ManifestFile

	kit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)

	manifestPath := filepath.Join(kitDir, catpkg, "Manifest")
	if utils.Exists(manifestPath) {
		// Manifest file could be not available for virtual packages.

		manifest, err = ParseManifest(manifestPath)
		if err != nil {
			return err
		}
	}

	pOpts := NewPortageResolverOpts()
	// Retrieve again all availables versions to avoid removing of the
	// files shared between multiple ebuilds
	atomsAvailables, err := m.TargetResolver.GetValidPackages(catpkg, pOpts)
	if err != nil {
		return err
	}

	// Exclude packages candidates from availables list
	validPackages := []*specs.RepoScanAtom{}
	for _, atom := range atomsAvailables {
		toSkip := false
		for idx := range candidates {
			if atom.Atom == candidates[idx].Atom {
				toSkip = true
				break
			}
		}

		if !toSkip {
			validPackages = append(validPackages, atom)
		}
	}
	// Free memory
	atomsAvailables = nil

	// Manifest file could
	manifest2Update := false
	files := []string{}
	cMsg := fmt.Sprintf("%s: Remove old versions\n", catpkg)

	// Remove candidates
	for _, candidate := range candidates {

		if manifest != nil || len(manifest.Files) > 0 {
			// Checking files to remove from Manifest

			// Check if the files of the candidates are used by existing packages
			for _, file := range candidate.Files {
				file2Skip := false
				for idx := range validPackages {
					if validPackages[idx].HasFile(file.Name) {
						file2Skip = true
						break
					}
				}

				if !file2Skip {
					manifest2Update = true
					manifest.RemoveFile(file.Name)
				}
			}
		}

		gp, _ := candidate.ToGentooPackage()

		ebuildPath := filepath.Join(kitDir, catpkg,
			fmt.Sprintf("%s.ebuild", gp.GetPF()))

		files = append(files, ebuildPath)

		err = os.Remove(ebuildPath)
		if err != nil {
			return err
		}

		cMsg += fmt.Sprintf("\n  * Removed v%s", gp.GetPVR())
	}

	if manifest2Update {
		err := manifest.Write(manifestPath)
		if err != nil {
			return err
		}

		files = append(files, manifestPath)
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

	if opts.PullRequest {
		prBranchName := fmt.Sprintf(
			"%s-%s",
			prBranchPurgePrefix,
			strings.ReplaceAll(strings.ReplaceAll(catpkg, ".", "_"),
				"/", "_"),
		)

		prBranchExists, err := BranchExists(kit.Url, prBranchName)
		if err != nil {
			return err
		}

		if prBranchExists {
			// PR is already been pushed.
			m.Logger.InfoC(fmt.Sprintf(
				"[%s] PR branch %s already present. Nothing to do.",
				catpkg, prBranchName))
			return nil
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
			Force:  true,
		}
		err := worktree.Checkout(&branchCoOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MergeBot) PushRemoves(mkit *specs.MergeKit, opts *MergeBotOpts,
	candidates4PN *map[string][]*specs.RepoScanAtom) error {

	var err error
	targetKit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), targetKit.Name)
	pushOpts := NewPushOptions()
	ctx := context.Background()

	if opts.PullRequest {
		// Setup github client for PR
		err = m.SetupGithubClient(ctx)
		if err != nil {
			return err
		}

		// Push purge branches
		for catpkg, candidates := range *candidates4PN {

			prBranchName := fmt.Sprintf(
				"%s-%s",
				prBranchPurgePrefix,
				strings.ReplaceAll(strings.ReplaceAll(catpkg, ".", "_"),
					"/", "_"),
			)

			err = PushBranch(kitDir, prBranchName, pushOpts)
			if err != nil {
				break
			}

			body := fmt.Sprintf(
				"Automatic clean of old versions of the package %s for branch %s by mark-bot\n\n",
				catpkg, targetKit.Branch)

			for idx := range candidates {
				gp, _ := candidates[idx].ToGentooPackage()
				body += fmt.Sprintf(
					" * v%s\n", gp.GetPVR())
			}

			pr, err := CreatePullRequest(m.GithubClient, ctx,
				// title
				fmt.Sprintf("mark-devkit: Purge %s", catpkg),
				// source branch
				prBranchName,
				// target branch
				targetKit.Branch,
				// body
				body,
				// github User
				opts.GithubUser,
				// github target repository
				targetKit.Name,
			)

			if err != nil {
				return err
			}

			m.Logger.Info(fmt.Sprintf("[%s] Created correctly PR: %s",
				catpkg, pr.GetHTMLURL()))
		}

	} else {
		err = Push(kitDir, pushOpts)
	}

	return err
}

func (m *MergeBot) SearchAtoms2Clean(mkit *specs.MergeKit, opts *MergeBotOpts) (*map[string]*specs.RepoScanAtom, error) {
	mapPkgs := make(map[string]*specs.RepoScanAtom, 0)

	for _, atom := range mkit.Target.Atoms {
		m.Logger.InfoC(fmt.Sprintf(":lollipop:[%s] Checking...",
			atom.Package))

		candidates, err := m.searchAtom2Clean(atom, mkit, opts)
		if err != nil {
			return nil, err
		}

		if len(candidates) > 0 {
			for _, candidate := range candidates {
				// TODO: add check if already set
				mapPkgs[candidate.Atom] = candidate
			}
		} else {
			m.Logger.Debug(fmt.Sprintf(
				":smiling_face_with_sunglasses:[%s] No versions to remove.",
				atom.Package))
		}
	}

	return &mapPkgs, nil
}

func (m *MergeBot) searchAtom2Clean(atom *specs.MergeKitAtom, mkit *specs.MergeKit,
	opts *MergeBotOpts) ([]*specs.RepoScanAtom, error) {
	ans := []*specs.RepoScanAtom{}

	pOpts := NewPortageResolverOpts()
	pOpts.Conditions = atom.Conditions

	availablesPkgs4Conditions, err := m.TargetResolver.GetValidPackages(atom.Package, pOpts)
	if err != nil {
		return ans, err
	}

	permittedVersions := 5
	if mkit.Target.AtomDefaults != nil {
		permittedVersions = *mkit.Target.AtomDefaults.MaxVersions
	}

	if atom.MaxVersions != nil {
		permittedVersions = *atom.MaxVersions
	}

	availablesPkgs := len(availablesPkgs4Conditions)
	// Decrement availables packages because the
	// last version could not be removed
	availablesPkgs--

	if availablesPkgs > permittedVersions {

		packages2Remove := len(availablesPkgs4Conditions) - permittedVersions

		aidx := 0

		for idx := range packages2Remove {

			for aidx < availablesPkgs {

				candidate := availablesPkgs4Conditions[aidx]
				gp, _ := gentoo.ParsePackageStr(candidate.Atom)

				m.Logger.DebugC(fmt.Sprintf(
					":lollipop:[%s] Checking candidate %s (%d/%d)...",
					atom.Package, gp.GetPF(), idx, packages2Remove))

				if len(atom.Versions) > 0 {
					toSkip := false
					// POST: check if the version is pinned
					for _, v := range atom.Versions {
						if v == gp.GetPVR() {
							toSkip = true
							break
						}
					}

					if toSkip {
						m.Logger.InfoC(fmt.Sprintf(
							":safety_pin:[%s] Version %s pinned. It will not be removed.",
							atom.Package, gp.GetPVR()))
						aidx++
						continue
					}
				}

				// POST: candidate valid for remove
				ans = append(ans, candidate)
				aidx++
			}

		}

	} // else nothing to do

	return ans, nil
}
