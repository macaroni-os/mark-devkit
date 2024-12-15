/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (m *MergeBot) MergeEclasses(mkit *specs.MergeKit, opts *MergeBotOpts) error {
	eMap := mkit.GetEclassesInclude()
	if eMap == nil {
		return nil
	}

	kit, _ := mkit.GetTargetKit()
	kitDir := filepath.Join(m.GetTargetDir(), kit.Name)

	eclassMap4Commit := make(map[string]string, 0)

	targetEclassdir := filepath.Join(kitDir, "eclass")
	if !utils.Exists(targetEclassdir) {
		err := helpers.EnsureDirWithoutIds(targetEclassdir, 0755)
		if err != nil {
			return err
		}
	}

	for k, rules := range *eMap {
		sKit := mkit.GetSourceKit(k)
		if sKit == nil {
			m.Logger.Warning(fmt.Sprintf(
				"No source kit %s found for eclasses.", k))
			continue
		}

		sourceKit := filepath.Join(m.GetSourcesDir(), k)

		err := m.mergeKitEclasses(rules, sourceKit, kitDir, mkit, opts, &eclassMap4Commit)
		if err != nil {
			return err
		}
	}

	m.Logger.Info(fmt.Sprintf(":dart:Found %d eclasses to add/updates.",
		len(eclassMap4Commit)))

	if len(eclassMap4Commit) > 0 {
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

		files := []string{}
		for _, file := range eclassMap4Commit {
			files = append(files, file)
		}

		cMsg := ""
		if m.IsANewBranch {
			cMsg = "Add eclasses"
		} else {
			cMsg = "Update/Add eclasses"
		}

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

func (m *MergeBot) mergeKitEclasses(rules []string, kitDir, targetKitDir string,
	mkit *specs.MergeKit, opts *MergeBotOpts, eclassMap *map[string]string) error {

	me := *eclassMap
	eclassDir := filepath.Join(kitDir, "eclass")
	targetEclassdir := filepath.Join(targetKitDir, "eclass")

	if !utils.Exists(eclassDir) {
		return nil
	}

	regexes := []*regexp.Regexp{}

	// Prepare regex
	for idx := range rules {
		regexes = append(regexes, regexp.MustCompile(rules[idx]))
	}

	entries, err := os.ReadDir(eclassDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		match := false

		for idx := range regexes {
			if regexes[idx] == nil {
				continue
			}

			if regexes[idx].MatchString(entry.Name()) {
				match = true
				break
			}
		}

		if !match {
			continue
		}

		sourceEclass := filepath.Join(eclassDir, entry.Name())
		targetEclass := filepath.Join(targetEclassdir, entry.Name())
		targetEclassMd5 := ""

		sourceEclassMd5, err := helpers.GetFileMd5(sourceEclass)
		if err != nil {
			return err
		}

		if utils.Exists(targetEclass) {
			targetEclassMd5, err = helpers.GetFileMd5(targetEclass)
			if err != nil {
				return err
			}
		}

		if sourceEclassMd5 != targetEclassMd5 {
			err = helpers.CopyFile(sourceEclass, targetEclass)
			if err != nil {
				return err
			}

			me[entry.Name()] = targetEclass
		}

	}

	return nil
}
