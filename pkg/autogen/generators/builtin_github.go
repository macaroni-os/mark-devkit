/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/google/go-github/v68/github"
)

type GithubGenerator struct {
	Client *github.Client
}

func NewGithubGenerator() *GithubGenerator {
	return &GithubGenerator{
		Client: nil,
	}
}

func (g *GithubGenerator) GetType() string {
	return specs.GeneratorBuiltinGitub
}

func (g *GithubGenerator) SetClient(c *github.Client) { g.Client = c }

func (g *GithubGenerator) GetAssets(atom *specs.AutogenAtom,
	release *github.RepositoryRelease,
	mapref *map[string]interface{}) ([]*specs.AutogenArtefact, error) {

	values := *mapref
	ans := []*specs.AutogenArtefact{}

	for _, asset := range atom.Assets {

		name, err := helpers.RenderContentWithTemplates(
			asset.Name,
			"", "", "asset.name", values, []string{},
		)
		if err != nil {
			return ans, err
		}

		r := regexp.MustCompile(asset.Matcher)
		if r == nil {
			return ans, fmt.Errorf("[%s] invalid regex on asset %s", atom.Name, asset.Name)
		}

		assetFound := false

		for idx := range release.Assets {
			if r.MatchString(release.Assets[idx].GetName()) {
				assetFound = true
				ans = append(ans, &specs.AutogenArtefact{
					SrcUri: []string{release.Assets[idx].GetBrowserDownloadURL()},
					Use:    asset.Use,
					Name:   name,
				})
				break
			}
		}

		if !assetFound {
			return ans, fmt.Errorf("[%s] no asset found for matcher %s", atom.Name, asset.Matcher)
		}

	}

	return ans, nil
}

func (g *GithubGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	values := *mapref

	var tag *github.RepositoryTag
	var release *github.RepositoryRelease
	var err error

	originalVersion, _ := values["original_version"].(string)

	// Set release metadata if present
	if _, present := values["releases"]; present {

		releases, _ := values["releases"].(map[string]*github.RepositoryRelease)
		release = releases[originalVersion]
		values["release"] = release

		tags, _ := values["tags"].(map[string]*github.RepositoryTag)
		tag = tags[release.GetTagName()]
		values["tag"] = tag

	} else {
		// Set only the tag
		tags, _ := values["tags"].(map[string]*github.RepositoryTag)
		tag = tags[originalVersion]
		values["tag"] = tag

	}

	values["sha"] = tag.Commit.GetSHA()

	delete(values, "releases")
	delete(values, "tags")
	delete(values, "versions")

	tarballName := atom.Tarball
	if tarballName == "" {
		tarballName = fmt.Sprintf("%s-%s.tar.gz", atom.Name, version)
	} else {
		tarballName, err = helpers.RenderContentWithTemplates(
			tarballName,
			"", "", "artefact.tarball", values, []string{},
		)
		if err != nil {
			return err
		}
	}

	artefacts := []*specs.AutogenArtefact{}

	if release != nil && (atom.HasAssets() || release.GetTarballURL() != "") {
		if atom.HasAssets() {
			artefacts, err = g.GetAssets(atom, release, mapref)
			if err != nil {
				return err
			}
		} else {
			artefacts = append(artefacts, &specs.AutogenArtefact{
				SrcUri: []string{release.GetTarballURL()},
				Name:   tarballName,
			})
		}

	} else {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{tag.GetTarballURL()},
			Name:   tarballName,
		})
	}

	values["artefacts"] = artefacts

	return nil
}

func (g *GithubGenerator) Process(atom *specs.AutogenAtom,
	def *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)
	ctx := context.Background()

	ghData := def.Clone()

	if atom.Github != nil {
		if atom.Github.User != "" {
			ghData.Github.User = atom.Github.User
		}
		if atom.Github.Repo != "" {
			ghData.Github.Repo = atom.Github.Repo
		}
		if atom.Github.Query != "" {
			ghData.Github.Query = atom.Github.Query
		}
		if atom.Github.PerPage != nil {
			ghData.Github.PerPage = atom.Github.PerPage
		}
		if atom.Github.Page != nil {
			ghData.Github.Page = atom.Github.Page
		}
		if atom.Github.NumPages != nil {
			ghData.Github.NumPages = atom.Github.NumPages
		}
	}

	if ghData.Github.Repo == "" {
		ghData.Github.Repo = atom.Name
	}

	if ghData.Github.Repo == "" {
		return nil, fmt.Errorf("no github repo defined for atom %s",
			atom.Name)
	}
	if ghData.Github.User == "" {
		return nil, fmt.Errorf("no github user defined for atom %s",
			atom.Name)
	}
	if ghData.Github.Query == "" {
		return nil, fmt.Errorf("no github query defined for atom %s",
			atom.Name)
	}
	if ghData.Github.Query != "releases" && ghData.Github.Query != "tags" {
		return nil, fmt.Errorf("github query with invalid query for atom %s",
			atom.Name)
	}

	var lopts *github.ListOptions = nil
	validTags := make(map[string]*github.RepositoryTag, 0)
	versions := []string{}

	if ghData.Github.Page != nil || ghData.Github.PerPage != nil {
		lopts = &github.ListOptions{}
		if ghData.Github.Page != nil {
			lopts.Page = *ghData.Github.Page
		} else {
			lopts.Page = 1
		}
		if ghData.Github.PerPage != nil {
			lopts.PerPage = *ghData.Github.PerPage
		}
	}

	if ghData.Github.Query == "tags" {
		tt := []*github.RepositoryTag{}

		if ghData.Github.NumPages != nil {
			for page := 1; page < *ghData.Github.NumPages; page++ {
				tags, resp, err := g.Client.Repositories.ListTags(
					ctx, ghData.Github.User, ghData.Github.Repo,
					lopts,
				)
				if err != nil {
					return nil, err
				}

				tt = append(tt, tags...)

				lopts.Page = resp.NextPage
				if lopts.Page > resp.LastPage {
					break
				}
			}

		} else {
			// POST: Read only one page.
			tags, _, err := g.Client.Repositories.ListTags(
				ctx, ghData.Github.User, ghData.Github.Repo, lopts,
			)
			if err != nil {
				return nil, err
			}

			tt = tags
		}

		for idx := range tt {
			version := tt[idx].GetName()

			if strings.HasPrefix(version, "v") {
				version = version[1:len(version)]
			}
			validTags[version] = tt[idx]

			versions = append(versions, version)
		}

	} else {
		rr := []*github.RepositoryRelease{}
		tagsMap := make(map[string]*github.RepositoryTag, 0)

		if ghData.Github.NumPages != nil {
			for page := 1; page < *ghData.Github.NumPages; page++ {
				releases, resp, err := g.Client.Repositories.ListReleases(
					ctx, ghData.Github.User, ghData.Github.Repo, lopts,
				)
				if err != nil {
					return nil, err
				}

				tags, resp, err := g.Client.Repositories.ListTags(
					ctx, ghData.Github.User, ghData.Github.Repo, lopts,
				)
				if err != nil {
					return nil, err
				}

				for i := range tags {
					tagsMap[tags[i].GetName()] = tags[i]
				}

				rr = append(rr, releases...)

				lopts.Page = resp.NextPage
				if lopts.Page > resp.LastPage {
					break
				}
			}

		} else {
			// POST: Read only one page.
			releases, _, err := g.Client.Repositories.ListReleases(
				ctx, ghData.Github.User, ghData.Github.Repo, lopts,
			)
			if err != nil {
				return nil, err
			}

			tags, _, err := g.Client.Repositories.ListTags(
				ctx, ghData.Github.User, ghData.Github.Repo, lopts,
			)
			if err != nil {
				return nil, err
			}

			for i := range tags {
				tagsMap[tags[i].GetName()] = tags[i]
			}

			rr = releases
		}

		validReleases := make(map[string]*github.RepositoryRelease, 0)

		for idx := range rr {
			if (rr[idx].Prerelease != nil && *rr[idx].Prerelease) ||
				(rr[idx].Draft != nil && *rr[idx].Draft) {
				continue
			}

			validTags[rr[idx].GetTagName()], _ = tagsMap[rr[idx].GetTagName()]
			version := rr[idx].GetName()

			if strings.HasPrefix(version, "v") {
				version = version[1:len(version)]
			}

			versions = append(versions, version)

			validReleases[version] = rr[idx]
		}

		ans["releases"] = validReleases

	}

	ans["versions"] = versions
	ans["tags"] = validTags
	ans["github_user"] = ghData.Github.User
	ans["github_repo"] = ghData.Github.Repo

	return &ans, nil
}
