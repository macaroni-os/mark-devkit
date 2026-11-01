/*
Copyright © 2024-2026 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	autogenart "github.com/macaroni-os/mark-devkit/pkg/autogen/artefacts"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	executor "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type ExtensionGolang struct {
	*ExtensionBase
}

type GoSum struct {
	Lines []GoSumRow
}

type GoSumRow struct {
	Module   string
	Version  string
	Checksum string
}

func NewGoSum(content string) *GoSum {
	lines := strings.Split(content, "\n")

	ans := &GoSum{
		Lines: []GoSumRow{},
	}

	escapeModStr := func(str string) string {
		var builder strings.Builder
		for _, ch := range str {
			if ch >= 'A' && ch <= 'Z' {
				builder.WriteByte('!')
				builder.WriteByte(byte(ch + 'a' - 'A'))
			} else {
				builder.WriteRune(ch)
			}
		}
		return builder.String()
	}

	for idx := range lines {
		// As described on https://golang.org/ref/mod#module-cache
		// we need to convert the upper case with !lower case
		line := escapeModStr(lines[idx])
		words := strings.Split(line, " ")
		if len(words) < 3 {
			continue
		}
		ans.Lines = append(ans.Lines, GoSumRow{
			Module:   strings.TrimSpace(words[0]),
			Version:  strings.TrimSpace(words[1]),
			Checksum: strings.TrimSpace(words[2]),
		})
	}

	return ans
}

func NewExtensionGolang(opts map[string]string) (*ExtensionGolang, error) {
	return &ExtensionGolang{
		ExtensionBase: &ExtensionBase{
			Opts: opts,
		}}, nil
}

func (e *ExtensionGolang) GetName() string { return specs.ExtensionGolang }

func (e *ExtensionGolang) Elaborate(restGuard *guard.RestGuard,
	atom, def *specs.AutogenAtom,
	mapref *map[string]interface{}) error {

	log := logger.GetDefaultLogger()
	values := *mapref

	// Create the temporary directory of the package
	workdir, _ := e.Opts["workdir"]
	downloadDir := e.Opts["download_dir"]
	if !filepath.IsAbs(downloadDir) {
		downloadDir, _ = filepath.Abs(downloadDir)
	}
	if !filepath.IsAbs(workdir) {
		workdir, _ = filepath.Abs(workdir)
	}

	pkgWorkDir := filepath.Join(workdir, "go-extension", atom.Name)

	err := os.MkdirAll(pkgWorkDir, os.ModePerm)
	if err != nil {
		return err
	}

	mirror := e.Opts["mirror"]
	if mirror == "" {
		mirror = "mirror://macaroni"
	}
	bundleIdentifier := e.Opts["bundle_identifier"]
	if bundleIdentifier == "" {
		bundleIdentifier = "mark-go-bundle"
	}
	bundleExtension := e.Opts["bundle_extension"]
	if bundleExtension == "" {
		bundleExtension = "xz"
	}
	// Fix typo on bundle_extension value.
	bundleExtension = strings.ReplaceAll(bundleExtension, ".", "")
	values["mirror"] = mirror

	// Ensure download dir. Could be not present the first time.
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Retrieve artefacts.
	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)
	// NOTE: I consider the first artefact the main for golang vendor generation.
	if len(artefacts) == 0 {
		log.DebugC(
			fmt.Sprintf(
				"[%s] No artefacts found for golang vendor generation. Nothing to do.",
				atom.Name,
			))
		return nil
	}

	art := artefacts[0]

	// Download the main artefacts to the download dir.
	log.DebugC(
		fmt.Sprintf("[%s] Downloading %s from url %s",
			atom.Name, art.Name, art.SrcUri[0],
		))
	repoFile, err := autogenart.DownloadArtefact(restGuard, atom,
		art.SrcUri[0], art.Name, downloadDir)
	if err != nil {
		return err
	}

	// Unpack the tarballs
	unpackDir := filepath.Join(pkgWorkDir, "unpack")
	log.Info(
		fmt.Sprintf(":factory:[%s] Extracting file %s",
			atom.Name, repoFile.Name,
		))
	err = e.unpackArtefact(downloadDir, unpackDir, repoFile,
		atom, def, mapref)
	if err != nil {
		return err
	}

	// Retrieve go.sum
	goSum, goSumDir, err := e.retrieveGoSum(atom, unpackDir)
	if err != nil {
		return err
	}

	// Merge the extra go.sum files, if any (multi module repositories)
	goSum, err = e.mergeExtraGoSums(atom, goSumDir, goSum)
	if err != nil {
		return err
	}

	// Download bundle files
	bundlesDir := filepath.Join(pkgWorkDir, bundleIdentifier+"-"+atom.Name)
	err = e.downloadBundles(restGuard, atom, goSum, bundlesDir)
	if err != nil {
		return err
	}

	// Create bundle tarball
	version, _ := values["version"].(string)
	sha, _ := values["sha"].(string)
	bundleTarball := ""
	if sha == "" {
		bundleTarball = fmt.Sprintf(
			"%s-%s-%s.tar.%s", atom.Name, version, bundleIdentifier, bundleExtension,
		)
	} else {
		bundleTarball = fmt.Sprintf(
			"%s-%s-%s-%s.tar.%s", atom.Name, version, bundleIdentifier,
			sha[0:7], bundleExtension,
		)
	}
	bundleArt, err := e.createBundleTarball(bundlesDir,
		filepath.Join(downloadDir, bundleTarball), workdir, atom.Name, mirror)
	if err != nil {
		return err
	}

	artefacts = append(artefacts, bundleArt)
	values["artefacts"] = artefacts

	e.cleanup(mapref)

	if !logger.GetDefaultLogger().Config.GetGeneral().Debug {
		defer os.RemoveAll(filepath.Join(workdir, "go-extension"))
	}

	return nil
}

func (e *ExtensionGolang) createBundleTarball(
	bundlesDir, bundleTarball, workDir, atomName, mirror string) (*specs.AutogenArtefact, error) {
	var err error
	log := logger.GetDefaultLogger()

	// Check instance
	config := tarf_specs.NewConfig(nil)
	if logger.GetDefaultLogger().Config.GetGeneral().Debug {
		config.GetLogging().Level = "info"
	} else {
		config.GetGeneral().Debug = false
		config.GetLogging().Level = "info"
	}

	tarformers := executor.NewTarFormers(config)
	s := tarf_specs.NewSpecFile()
	// We don't need to keep the original permission of the files
	// and owner.
	s.SameOwner = false
	s.SameChtimes = false
	s.Writer = tarf_specs.NewWriter()
	s.Writer.ArchiveDirs = []string{bundlesDir}

	s.RenamePath = []tarf_specs.RenameRule{
		{
			Source: filepath.Join(workDir, "go-extension", atomName),
			Dest:   ".",
		},
	}
	tarfOpts := tools.NewTarCompressionOpts(true)
	defer tarfOpts.Close()

	err = tools.PrepareTarWriter(bundleTarball, tarfOpts)
	if err != nil {
		return nil, fmt.Errorf(
			"error on prepare writer: %s",
			err.Error())
	}

	if tarfOpts.CompressWriter != nil {
		tarformers.SetWriter(tarfOpts.CompressWriter)
	} else {
		tarformers.SetWriter(tarfOpts.FileWriter)
	}

	fileName := filepath.Base(bundleTarball)
	log.Debug(fmt.Sprintf("[%s] Creating tarball %s...",
		atomName, fileName))

	err = tarformers.RunTaskWriter(s)
	if err != nil {
		return nil, fmt.Errorf(
			"srror on create tarball %s: %s",
			fileName, err.Error())
	}

	artUri := fmt.Sprintf("%s/%s", mirror, fileName)

	local := true
	return &specs.AutogenArtefact{
		SrcUri: []string{artUri},
		Name:   fileName,
		Local:  &local,
	}, nil
}

func (e *ExtensionGolang) downloadBundles(restGuard *guard.RestGuard,
	atom *specs.AutogenAtom, goSum *GoSum,
	bundlesDir string) error {
	log := logger.GetDefaultLogger()
	var err error

	// Create bundle dir
	err = os.MkdirAll(bundlesDir, os.ModePerm)
	if err != nil {
		return err
	}

	for idx := range goSum.Lines {
		moduleExt := "zip"
		if strings.HasSuffix(goSum.Lines[idx].Version, "go.mod") {
			moduleExt = "mod"
		}
		version := strings.Split(goSum.Lines[idx].Version, "/")[0]
		moduleUri := fmt.Sprintf("%s/@v/%s.%s",
			goSum.Lines[idx].Module, version, moduleExt)
		url := fmt.Sprintf("https://proxy.golang.org/%s", moduleUri)

		// Replace / with %2F
		bundle := strings.ReplaceAll(moduleUri, "/", "%2F")

		log.Debug(fmt.Sprintf("[%s] Downloading bundle %s %s at %s...",
			atom.Name, goSum.Lines[idx].Module, goSum.Lines[idx].Version,
			url))

		_, err = autogenart.DownloadArtefact(
			restGuard, atom, url,
			bundle, bundlesDir)
		if err != nil {
			return err
		}

	}

	return nil
}

// retrieveGoSum retrieves the main go.sum file of the unpacked artefact and
// return it parsed, with its directory (needed to resolve the extra go.sum
// paths) and any error.
func (e *ExtensionGolang) retrieveGoSum(atom *specs.AutogenAtom,
	targetDir string) (*GoSum, string, error) {

	unpackDirPrefix, _ := e.Opts["unpack_srcdir_prefix"]
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, "", err
	}

	goSumPath := ""

	for _, entry := range entries {

		if !entry.IsDir() {
			continue
		}

		goSumPath = filepath.Join(targetDir, entry.Name(), "go.sum")
		if unpackDirPrefix != "" && strings.HasPrefix(entry.Name(), unpackDirPrefix) {
			if !utils.Exists(goSumPath) {
				goSumPath = ""
				break
			}
		}

		if utils.Exists(goSumPath) {
			break
		}
		goSumPath = ""
	}

	if goSumPath == "" {
		return nil, "", fmt.Errorf("go.sum file not found")
	}

	data, err := os.ReadFile(goSumPath)
	if err != nil {
		return nil, "", fmt.Errorf("error on read file go.sum: %s", err.Error())
	}

	return NewGoSum(string(data)), filepath.Dir(goSumPath), nil
}

// mergeExtraGoSums merges the additional go.sum files listed in the
// extra_go_sums option into goSum. The option is a comma separated list of
// paths, relative to the directory of the main go.sum; it is needed to support
// repositories with multiple Go modules, where the nested modules keep their
// own go.sum (e.g. the collector/ module of grafana/alloy).
func (e *ExtensionGolang) mergeExtraGoSums(atom *specs.AutogenAtom,
	goSumDir string, goSum *GoSum) (*GoSum, error) {
	log := logger.GetDefaultLogger()

	extraFiles := e.Opts["extra_go_sums"]
	if strings.TrimSpace(extraFiles) == "" {
		return goSum, nil
	}

	for _, extraFile := range strings.Split(extraFiles, ",") {
		extraFile = strings.TrimSpace(extraFile)
		if extraFile == "" {
			continue
		}

		extraPath := extraFile
		if !filepath.IsAbs(extraPath) {
			extraPath = filepath.Join(goSumDir, extraFile)
		}

		if !utils.Exists(extraPath) {
			return nil, fmt.Errorf("error on read extra go.sum: file %s not found",
				extraFile)
		}

		data, err := os.ReadFile(extraPath)
		if err != nil {
			return nil, fmt.Errorf("error on read file %s: %s",
				extraFile, err.Error())
		}

		added := mergeGoSums(goSum, NewGoSum(string(data)))
		log.Debug(fmt.Sprintf("[%s] Merged %d rows from extra go.sum %s",
			atom.Name, added, extraFile))
	}

	return goSum, nil
}

// mergeGoSums appends to main the rows of extra not already known, matched by
// module@version, to avoid downloading the same bundle twice, and return the
// number of rows added.
func mergeGoSums(main, extra *GoSum) int {
	known := make(map[string]bool, len(main.Lines))
	for idx := range main.Lines {
		known[fmt.Sprintf("%s@%s", main.Lines[idx].Module,
			main.Lines[idx].Version)] = true
	}

	added := 0
	for idx := range extra.Lines {
		key := fmt.Sprintf("%s@%s", extra.Lines[idx].Module,
			extra.Lines[idx].Version)
		if known[key] {
			continue
		}
		known[key] = true
		main.Lines = append(main.Lines, extra.Lines[idx])
		added++
	}

	return added
}
