/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"

	tmpleng "github.com/macaroni-os/mark-devkit/pkg/autogen/tmpl-engines"
	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func (a *AutogenBot) GeneratePackageOnStaging(mkit *specs.MergeKit,
	aspec *specs.AutogenSpec, atom, def *specs.AutogenAtom,
	mapref *map[string]interface{},
	tmplEngine tmpleng.TemplateEngine) (*specs.RepoScanAtom, error) {

	values := *mapref
	category := atom.GetCategory(def)
	targetKit, _ := mkit.GetTargetKit()
	pkgDirStaging := filepath.Join(a.GetSourcesDir(), targetKit.Name,
		category, atom.Name)

	pn, _ := values["pn"].(string)
	version, _ := values["version"].(string)
	artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)

	ans := &specs.RepoScanAtom{
		Atom:     fmt.Sprintf("%s/%s-%s", category, atom.Name, version),
		Category: category,
		Package:  pn,
		Revision: "0",
		CatPkg:   fmt.Sprintf("%s/%s", category, atom.Name),
		Kit:      targetKit.Name,
		Branch:   targetKit.Branch,
		Files:    []specs.RepoScanFile{},
		// We need to have KEYWORDS for merge check
		Metadata: map[string]string{
			"KEYWORDS": "*",
		},
	}

	// Prepare package dir
	err := helpers.EnsureDirWithoutIds(pkgDirStaging, 0755)
	if err != nil {
		return nil, err
	}

	if len(artefacts) > 0 && (atom.IgnoreArtefacts == nil || !*atom.IgnoreArtefacts) {
		// Download tarballs
		for idx, art := range artefacts {
			filename := art.Name
			if art.Name == "" {
				filename = fmt.Sprintf("%s-%s.tar.gz", pn, version)
			}
			repoFile, err := a.downloadArtefact(art.SrcUri[0], art.Name)
			if err != nil {
				return nil, err
			}

			ans.Files = append(ans.Files, *repoFile)

			if idx == 0 {
				// Set the first artefact as src_uri
				values["src_uri"] = fmt.Sprintf("%s -> %s", art.SrcUri[0], filename)
			}
		}

		manifestPath := filepath.Join(pkgDirStaging, "Manifest")
		var manifest *kit.ManifestFile
		// Check if exists already a Manifest with other files.
		// For example the same package is been defined multiple time
		// with multiple selector.
		if utils.Exists(manifestPath) {
			manifest, err = kit.ParseManifest(manifestPath)
			if err != nil {
				return nil, err
			}
			manifest.AddFiles(ans.Files)
		} else {
			manifest = kit.NewManifestFile(ans.Files)
		}

		// Generate Manifest
		err := manifest.Write(manifestPath)
		if err != nil {
			return nil, err
		}
	}

	ebuildPath := filepath.Join(pkgDirStaging, fmt.Sprintf("%s-%s.ebuild", pn, version))
	err = tmplEngine.Render(aspec, atom, def, mapref, ebuildPath)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

func (a *AutogenBot) downloadArtefact(atomUrl, tarballName string) (*specs.RepoScanFile, error) {

	ans := &specs.RepoScanFile{
		SrcUri: []string{atomUrl},
		Name:   tarballName,
		Hashes: make(map[string]string, 0),
	}

	uri, err := url.Parse(atomUrl)
	if err != nil {
		return nil, err
	}

	ssl := false

	switch uri.Scheme {
	case "https":
		ssl = true
	default:
		ssl = false
	}

	if uri.Scheme == "ftp" {
		return nil, fmt.Errorf("Not yet implemented!")
	} else {

		node := guard_specs.NewRestNode(uri.Host,
			uri.Host+path.Dir(uri.Path), ssl)

		resource := path.Base(uri.Path)

		service := guard_specs.NewRestService(uri.Host)
		service.Retries = 3
		service.AddNode(node)

		t := service.GetTicket()
		defer t.Rip()

		_, err := a.RestGuard.CreateRequest(t, "GET", "/"+resource)
		if err != nil {
			return nil, err
		}

		downloadedFilePath := filepath.Join(a.GetDownloadDir(), tarballName)

		artefact, err := a.RestGuard.DoDownload(t, downloadedFilePath)
		if err != nil {
			if t.Response != nil {
				return nil, fmt.Errorf("%s - %s", err.Error(), t.Response.Status)
			} else {
				return nil, fmt.Errorf("%s", err.Error())
			}
		}

		ans.Hashes["sha512"] = artefact.Sha512
		ans.Hashes["blake2b"] = artefact.Blake2b
		ans.Size = fmt.Sprintf("%d", artefact.Size)
	}

	return ans, nil
}

func (a *AutogenBot) isVersion2Add(atom, def *specs.AutogenAtom,
	version string) (bool, error) {

	if a.MergeBot.TargetKitIsANewBranch() {
		return true, nil
	}

	catpkg := fmt.Sprintf("%s/%s", atom.GetCategory(def), atom.Name)

	if !a.MergeBot.TargetResolver.IsPresentPackage(catpkg) {
		return true, nil
	}

	pOpts := kit.NewPortageResolverOpts()
	// Retrieve all availables version in order to return true
	// and check for differences.
	atomsAvailables, err := a.MergeBot.TargetResolver.GetValidPackages(catpkg, pOpts)
	if err != nil {
		return false, err
	}

	// Prepare GentooPackage of the selected version
	gpkg, err := gentoo.ParsePackageStr(fmt.Sprintf("%s-%s", catpkg, version))
	if err != nil {
		return false, err
	}

	toAdd := true
	for idx := range atomsAvailables {
		agpg, _ := atomsAvailables[idx].ToGentooPackage()
		if equal, _ := agpg.Equal(gpkg); equal {
			return true, nil
		}

		if toskip, _ := agpg.GreaterThan(gpkg); toskip {
			toAdd = false
		}
	}

	// Free memory
	atomsAvailables = nil

	return toAdd, nil
}
