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

		commitHash, err := m.commitFiles(kitDir, []string{layoutConf},
			cMsg, opts, worktree)
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
