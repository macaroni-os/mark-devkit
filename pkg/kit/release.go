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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"
)

const (
	reposConfDefault = `[DEFAULT]
main-repo = %s
`
	reposConf = `[%s]
location = %s
auto-sync = no
priority = %d
`
)

type ReleaseBot struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	IsANewBranch bool
	WorkDir      string
	MetaRepoPath string

	files2Add []string
	files2Del []string
}

type ReleaseOpts struct {
	DryRun  bool
	Push    bool
	Verbose bool

	GitDeepFetch int

	SignatureName  string
	SignatureEmail string
}

func NewReleaseOpts() *ReleaseOpts {
	return &ReleaseOpts{
		DryRun:       false,
		Push:         false,
		Verbose:      false,
		GitDeepFetch: 10,
	}
}

func NewReleaseBot(c *specs.MarkDevkitConfig) *ReleaseBot {
	return &ReleaseBot{
		Config:       c,
		Logger:       log.GetDefaultLogger(),
		WorkDir:      "./workdir",
		MetaRepoPath: "/var/git/meta-repo",
	}
}

func (r *ReleaseBot) GetSourcesDir() string {
	return filepath.Join(r.WorkDir, "sources")
}

func (r *ReleaseBot) GetTargetDir() string {
	return filepath.Join(r.WorkDir, "dest")
}

func (r *ReleaseBot) GetRepoPath(kit string) string {
	return filepath.Join(r.MetaRepoPath, "kits/", kit)
}

func (r *ReleaseBot) SetWorkDir(d string) { r.WorkDir = d }

func (r *ReleaseBot) Run(specfile string, opts *ReleaseOpts) error {
	// Load KitRelease specs
	release := specs.NewKitReleaseSpec()

	// Load specfile with release bump data.
	err := release.LoadFile(specfile)
	if err != nil {
		return err
	}

	// Clone source kits
	err = r.cloneSourcesKit(release, opts)
	if err != nil {
		return err
	}

	// Clone target kit
	err = r.cloneTargetKit(release, opts)
	if err != nil {
		return err
	}

	// Generate meta-repo/repos.conf/ files
	err = r.prepareReposConfDir(release, opts)
	if err != nil {
		return err
	}

	// Generate meta-repo/metadata/ files
	err = r.prepareMetadataDir(release, opts)
	if err != nil {
		return err
	}

	if len(r.files2Add) > 0 || len(r.files2Del) > 0 {
		// Run commit
		err = r.bumpRelease(release, opts)
		if err != nil {
			return err
		}
	}

	if opts.Push {
		kit := &release.Release.Target
		repoDir := filepath.Join(r.GetTargetDir(), kit.Name)

		pushOpts := NewPushOptions()
		err = Push(repoDir, pushOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReleaseBot) bumpRelease(release *specs.KitReleaseSpec,
	opts *ReleaseOpts) error {

	kit := &release.Release.Target
	repoDir := filepath.Join(r.GetTargetDir(), kit.Name)

	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	files := append(r.files2Add, r.files2Del...)

	cMsg := "Bump kits updates"
	if r.IsANewBranch {
		cMsg = "Bump first release :champagne:!"
	}

	commitHash, err := r.commitFiles(repoDir, files, cMsg, opts, worktree)
	if err != nil {
		return err
	}

	if opts.Verbose {
		commit, _ := repo.CommitObject(commitHash)
		r.Logger.InfoC(fmt.Sprintf("%s", commit))
	}

	return nil
}

func (r *ReleaseBot) commitFiles(kitDir string, files []string,
	commitMessage string, opts *ReleaseOpts,
	worktree *git.Worktree) (plumbing.Hash, error) {

	for _, file := range files {
		// Drop kitDir prefix
		f := file[len(kitDir)+1 : len(file)]
		_, err := worktree.Add(f)
		if err != nil {
			return plumbing.ZeroHash, err
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

func (r *ReleaseBot) prepareReposConfDir(release *specs.KitReleaseSpec,
	opts *ReleaseOpts) error {
	reposConfFiles := make(map[string]string, 0)
	kit := &release.Release.Target
	reposConfDir := filepath.Join(r.GetTargetDir(), kit.Name, "repos.conf")

	if utils.Exists(reposConfDir) {
		// Read the existing files and content
		entries, err := os.ReadDir(reposConfDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			file := filepath.Join(reposConfDir, entry.Name())
			md5Entry, err := helpers.GetFileMd5(file)

			// The default file is mandatory and contains
			// the info about what is the main-repo kit.
			if entry.Name() == "default" {
				reposConfFiles[entry.Name()] = md5Entry

			} else {

				// Check if the file name is related to an existing
				// kit.
				skit := release.GetSourceKit(entry.Name())
				if skit == nil {
					// POST: kit not defined. I will drop the file
					r.files2Del = append(r.files2Del, file)
					err = os.Remove(file)
					if err != nil {
						return err
					}
				} else {
					// POST: kit exists. Add file for second step.
					reposConfFiles[entry.Name()] = md5Entry

				}
			}
		}

	} else {
		err := helpers.EnsureDirWithoutIds(reposConfDir, 0755)
		if err != nil {
			return err
		}
	}

	// Iterate for every kit
	for _, skit := range release.Release.Sources {
		reposKitFile := filepath.Join(reposConfDir, skit.Name)

		content := fmt.Sprintf(reposConf,
			skit.Name,
			r.GetRepoPath(skit.Name),
			skit.GetPriority(),
		)

		newMd5 := fmt.Sprintf("%x", md5.Sum([]byte(content)))

		md5ExistingFile, present := reposConfFiles[skit.Name]
		if !present || md5ExistingFile != newMd5 {
			err := os.WriteFile(reposKitFile, []byte(content), 0644)
			if err != nil {
				return err
			}

			r.files2Add = append(r.files2Add, reposKitFile)
		}
	}

	// Check if I need create the default file

	defaultContent := fmt.Sprintf(reposConfDefault,
		release.Release.GetMainKit())
	defaultMd5 := fmt.Sprintf("%x", md5.Sum([]byte(defaultContent)))

	defaultFile := filepath.Join(reposConfDir, "default")
	md5ExistingFile, present := reposConfFiles["default"]
	if !present || md5ExistingFile != defaultMd5 {
		err := os.WriteFile(defaultFile, []byte(defaultContent), 0644)
		if err != nil {
			return err
		}

		r.files2Add = append(r.files2Add, defaultFile)
	}

	return nil
}

func (r *ReleaseBot) prepareMetadataDir(release *specs.KitReleaseSpec,
	opts *ReleaseOpts) error {

	kit := &release.Release.Target
	metadataDir := filepath.Join(r.GetTargetDir(), kit.Name, "metadata")

	if !utils.Exists(metadataDir) {
		err := helpers.EnsureDirWithoutIds(metadataDir, 0755)
		if err != nil {
			return err
		}
	}

	egoMap := make(map[string]string, 0)
	egoMap["ego"] = "2.7.2"

	version := &specs.MetaReleaseInfo{
		Required: []map[string]string{egoMap},
		Version:  11,
	}

	versionData, err := version.Json()
	if err != nil {
		return err
	}

	// Generate the version.json file
	sourceversionJsonMd5 := fmt.Sprintf("%x", md5.Sum(versionData))
	targetversionJsonMd5 := ""
	targetversionFile := filepath.Join(metadataDir, "version.json")

	if utils.Exists(targetversionFile) {
		targetversionJsonMd5, err = helpers.GetFileMd5(targetversionFile)
		if err != nil {
			return err
		}
	}

	if sourceversionJsonMd5 != targetversionJsonMd5 {
		err = os.WriteFile(targetversionFile, versionData, 0644)
		if err != nil {
			return err
		}
		r.files2Add = append(r.files2Add, targetversionFile)
	}

	// Generate the kit-info.json
	kitInfoFile := filepath.Join(metadataDir, "kit-info.json")
	// Prepare the MetaKitInfo structure
	info := &specs.MetaKitInfo{
		ReleaseInfo: version,
		KitSettings: make(map[string]specs.MetaKitSetting, 0),
		ReleaseDefs: make(map[string][]string, 0),
	}

	// Iterate for every kit
	for _, skit := range release.Release.Sources {
		info.KitOrder = append(info.KitOrder, skit.Name)
		kitSetting := specs.MetaKitSetting{
			Stability: map[string]string{
				skit.Branch: "prime",
			},
			Type: "auto",
		}

		info.KitSettings[skit.Name] = kitSetting
		info.ReleaseDefs[skit.Name] = []string{skit.Branch}
	}

	infoData, err := info.Json()
	if err != nil {
		return err
	}
	infoMd5 := fmt.Sprintf("%x", md5.Sum(infoData))
	targetInfoMd5 := ""

	if utils.Exists(kitInfoFile) {
		targetInfoMd5, err = helpers.GetFileMd5(kitInfoFile)
		if err != nil {
			return err
		}
	}

	if infoMd5 != targetInfoMd5 {
		err = os.WriteFile(kitInfoFile, infoData, 0644)
		if err != nil {
			return err
		}
		r.files2Add = append(r.files2Add, kitInfoFile)
	}

	// Generate the kit-sha1.json
	kitSha1File := filepath.Join(metadataDir, "kit-sha1.json")
	kitSha1 := &specs.MetaKitSha1{
		Kits: make(map[string]map[string]interface{}, 0),
	}

	// Iterate for every kit
	for _, skit := range release.Release.Sources {
		if skit.Depth != nil {
			kitSha1.Kits[skit.Name] = map[string]interface{}{
				skit.Branch: &specs.MetaKitShaValue{
					Sha1:  skit.CommitSha1,
					Depth: skit.Depth,
				},
			}
		} else {
			kitSha1.Kits[skit.Name] = map[string]interface{}{
				skit.Branch: skit.CommitSha1,
			}
		}
	}

	kitSha1Data, err := kitSha1.Json()
	kitSha1Md5 := fmt.Sprintf("%x", md5.Sum(kitSha1Data))
	targetKitSha1Md5 := ""

	if utils.Exists(kitSha1File) {
		targetKitSha1Md5, err = helpers.GetFileMd5(kitSha1File)
		if err != nil {
			return err
		}
	}

	if kitSha1Md5 != targetKitSha1Md5 {
		err = os.WriteFile(kitSha1File, kitSha1Data, 0644)
		if err != nil {
			return err
		}
		r.files2Add = append(r.files2Add, kitSha1File)
	}

	return nil
}

func (r *ReleaseBot) cloneSourcesKit(release *specs.KitReleaseSpec,
	opts *ReleaseOpts) error {
	gitOpts := &CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			SingleBranch: true,
			RemoteName:   "origin",
			Depth:        5,
		},
		Verbose: opts.Verbose,
		// Always generate summary report
		Summary: true,
		Results: []*specs.ReposcanKit{},
	}

	analysis := &specs.ReposcanAnalysis{
		Kits: release.Release.Sources,
	}

	// Clone sources kit
	err := CloneKits(analysis, r.GetSourcesDir(), gitOpts)
	if err != nil {
		return err
	}

	// Register the Results as sha1 of the kits.
	mRes := make(map[string]*specs.ReposcanKit, 0)
	for _, kit := range gitOpts.Results {
		mRes[kit.Name] = kit
	}

	for idx := range release.Release.Sources {
		kitName := release.Release.Sources[idx].Name
		release.Release.Sources[idx].CommitSha1 = mRes[kitName].CommitSha1
	}

	return nil
}

func (r *ReleaseBot) cloneTargetKit(release *specs.KitReleaseSpec, opts *ReleaseOpts) error {
	kit, err := release.Release.GetTargetKit()
	if err != nil {
		return err
	}

	r.Logger.Info(fmt.Sprintf(":factory:[%s] Cloning target kit...", kit.Name))
	kitDir := filepath.Join(r.GetTargetDir(), kit.Name)

	// Check if the repository branch exists.
	existsBranch, err := BranchExists(kit.Url, kit.Branch)
	if err != nil {
		return err
	}

	gitOpts := &CloneOptions{
		GitCloneOptions: &git.CloneOptions{
			RemoteName: "origin",
			Depth:      opts.GitDeepFetch,
		},
		Verbose: opts.Verbose,
		// Always generate summary report
		Summary: true,
		Results: []*specs.ReposcanKit{},

		SignatureName:  opts.SignatureName,
		SignatureEmail: opts.SignatureEmail,
	}

	if !existsBranch {
		r.Logger.InfoC(fmt.Sprintf("Branch %s doesn't exists. Creating the branch.",
			kit.Branch))

		r.IsANewBranch = true

		err = CloneAndCreateBranch(kit, kitDir, gitOpts)
	} else {
		gitOpts.GitCloneOptions.SingleBranch = true

		err = Clone(kit, kitDir, gitOpts)
	}

	return err
}
