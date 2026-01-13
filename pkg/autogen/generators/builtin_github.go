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
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/google/go-github/v74/github"
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

		if asset.Url != "" {

			assetUrl, err := helpers.RenderContentWithTemplates(
				asset.Url,
				"", "", "asset.url", values, []string{},
			)
			if err != nil {
				return ans, err
			}

			ans = append(ans, &specs.AutogenArtefact{
				SrcUri: []string{assetUrl},
				Use:    asset.Use,
				Name:   name,
			})

		} else {

			if release == nil {
				return ans, fmt.Errorf("matcher on asset is not permitted without using query release.")
			}

			matcher, err := helpers.RenderContentWithTemplates(
				asset.Matcher,
				"", "", "asset.matcher", values, []string{},
			)
			if err != nil {
				return ans, err
			}

			r := regexp.MustCompile(matcher)
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
				return ans, fmt.Errorf("[%s] no asset found for matcher %s", atom.Name, matcher)
			}
		}

	}

	return ans, nil
}

func (g *GithubGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {
	log := logger.GetDefaultLogger()

	values := *mapref

	var tag *github.RepositoryTag
	var release *github.RepositoryRelease
	var err error
	var sha string

	originalVersion, _ := values["original_version"].(string)

	// Set release metadata if present
	if _, present := values["releases"]; present {

		releases, _ := values["releases"].(map[string]*github.RepositoryRelease)
		release = releases[originalVersion]
		values["release"] = release

		tags, _ := values["tags"].(map[string]*github.RepositoryTag)
		tag = tags[release.GetTagName()]
		if tag == nil {
			// Try to search tag name with v. Really i hate this.
			tag = tags["v"+release.GetTagName()]
		}
		values["tag"] = tag

	} else {
		// Set only the tag
		tags, _ := values["tags"].(map[string]*github.RepositoryTag)
		tag = tags[originalVersion]
		values["tag"] = tag

	}

	if tag != nil {
		sha = tag.Commit.GetSHA()
	} else if atom.GithubIgnoreTags() && release != nil {
		sha = release.GetTargetCommitish()
	} else {
		log.Warning(fmt.Sprintf(
			"[%s] tag object not found for version %s (%s). Check if you need increase page elements and/or number of pages.",
			atom.Name, originalVersion, version))
	}

	values["sha"] = sha
	delete(values, "releases")
	delete(values, "tags")
	delete(values, "versions")

	tarballName := atom.Tarball
	if tarballName == "" {
		// Using sha at the end to correctly catch issues
		// with retag done on upstream repo
		if sha != "" && len(sha) > 7 {
			fmt.Println("SHA ", sha)
			tarballName = fmt.Sprintf("%s-%s-%s.tar.gz", atom.Name, version,
				sha[0:7])
		} else {
			tarballName = fmt.Sprintf("%s-%s.tar.gz", atom.Name, version)
		}
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

	if atom.HasAssets() {
		artefacts, err = g.GetAssets(atom, release, mapref)
		if err != nil {
			return err
		}
	} else if release != nil && release.GetTarballURL() != "" {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{release.GetTarballURL()},
			Name:   tarballName,
		})

	} else if tag != nil {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{tag.GetTarballURL()},
			Name:   tarballName,
		})
	}

	values["artefacts"] = artefacts

	if sha != "" && len(sha) > 7 {
		values["pkg_basedir"] = fmt.Sprintf("%s-%s-%s",
			atom.Github.User, atom.Github.Repo, sha[0:7],
		)
	} else {
		values["pkg_basedir"] = fmt.Sprintf("%s-%s-%s",
			atom.Github.User, atom.Github.Repo, originalVersion,
		)
	}

	return nil
}

func (g *GithubGenerator) getTags(atom *specs.AutogenAtom,
	lopts *github.ListOptions) ([]*github.RepositoryTag, error) {
	tt := []*github.RepositoryTag{}
	ctx := context.Background()
	log := logger.GetDefaultLogger()

	if atom.Github.NumPages != nil {
		for page := 1; page < *atom.Github.NumPages; page++ {
			tags, resp, err := g.Client.Repositories.ListTags(
				ctx, atom.Github.User, atom.Github.Repo,
				lopts,
			)
			if err != nil {
				return tt, err
			}

			tt = append(tt, tags...)

			if log.Config.GetGeneral().Debug {
				for _, t := range tags {
					log.Debug(fmt.Sprintf(
						"[%s] Found tag %s at page %d.",
						atom.Name, strings.ReplaceAll(t.GetName(), "\n", ""), page))
				}
			}

			lopts.Page = resp.NextPage
			if lopts.Page > resp.LastPage {
				break
			}
		}

	} else {
		// POST: Read only one page.
		tags, _, err := g.Client.Repositories.ListTags(
			ctx, atom.Github.User, atom.Github.Repo, lopts,
		)
		if err != nil {
			return tt, err
		}

		if log.Config.GetGeneral().Debug {
			for _, t := range tags {
				log.Debug(fmt.Sprintf(
					"[%s] Found tag %s.",
					atom.Name, strings.ReplaceAll(t.GetName(), "\n", "")))
			}
		}

		tt = tags
	}

	return tt, nil
}

func (g *GithubGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)
	ctx := context.Background()
	log := logger.GetDefaultLogger()
	var matchRegex *regexp.Regexp

	// Use atom.Name as default value for github repo and user if not defined.
	if atom.Github.Repo == "" {
		atom.Github.Repo = atom.Name
	}

	if atom.Github.User == "" {
		atom.Github.User = atom.Name
	}

	// Using release as default values for github if not defined.
	if atom.Github.Query == "" {
		atom.Github.Query = "releases"
	}

	if atom.Github.Query != "releases" && atom.Github.Query != "tags" {
		return nil, fmt.Errorf("github query with invalid query for atom %s",
			atom.Name)
	}

	var lopts *github.ListOptions = nil
	validTags := make(map[string]*github.RepositoryTag, 0)
	versions := []string{}
	r := regexp.MustCompile("^v[0-9].*")

	if atom.Github.Page != nil || atom.Github.PerPage != nil || atom.Github.NumPages != nil {
		lopts = &github.ListOptions{}
		if atom.Github.Page != nil {
			lopts.Page = *atom.Github.Page
		} else {
			lopts.Page = 1
		}
		if atom.Github.NumPages == nil {
			npages := 1
			atom.Github.NumPages = &npages
		}
		if atom.Github.PerPage != nil {
			lopts.PerPage = *atom.Github.PerPage
		}
	}

	if atom.Github.Query == "tags" {

		tt, err := g.getTags(atom, lopts)
		if err != nil {
			return nil, err
		}

		if atom.Github.Match != "" {
			matchRegex = regexp.MustCompile(atom.Github.Match)
			if matchRegex == nil {
				return nil, fmt.Errorf("invalid regex match string for atom %s",
					atom.Name)
			}
		}

		for idx := range tt {
			version := tt[idx].GetName()

			if matchRegex != nil && (!matchRegex.MatchString(version)) {
				log.Debug(fmt.Sprintf(
					"[%s] Tag %s doesn't match with regex. Ignore it.",
					atom.Name, version))
				continue
			}

			// Exclude v from tag name if related to a version
			if r.MatchString(version) {
				version = version[1:]
			}
			validTags[version] = tt[idx]

			versions = append(versions, version)
		}

	} else {
		// POST: query == releases

		if atom.Github.Match != "" {
			matchRegex = regexp.MustCompile(atom.Github.Match)
			if matchRegex == nil {
				return nil, fmt.Errorf("invalid regex match string for atom %s",
					atom.Name)
			}
		}

		rr := []*github.RepositoryRelease{}
		tagsMap := make(map[string]*github.RepositoryTag, 0)

		if !atom.GithubIgnoreTags() {
			tags, err := g.getTags(atom, lopts)
			if err != nil {
				return nil, err
			}

			for i := range tags {
				tagsMap[tags[i].GetName()] = tags[i]
			}
		}

		if atom.Github.NumPages != nil {
			lopts.Page = 1
			for page := 1; page < *atom.Github.NumPages; page++ {
				releases, resp, err := g.Client.Repositories.ListReleases(
					ctx, atom.Github.User, atom.Github.Repo, lopts,
				)
				if err != nil {
					return nil, err
				}

				if log.Config.GetGeneral().Debug {
					for _, r := range releases {
						log.Debug(fmt.Sprintf(
							"[%s] Found release %s at page %d (%d).",
							atom.Name, strings.ReplaceAll(r.GetName(), "\n", ""), page, resp.LastPage))
					}
				}
				rr = append(rr, releases...)

				// NextPage contains the number of page not the next!
				lopts.Page++
				if lopts.Page > resp.LastPage {
					break
				}
			}

		} else {
			// POST: Read only one page.
			releases, _, err := g.Client.Repositories.ListReleases(
				ctx, atom.Github.User, atom.Github.Repo, lopts,
			)
			if err != nil {
				return nil, err
			}

			rr = releases
		}

		validReleases := make(map[string]*github.RepositoryRelease, 0)
		var present bool

		for idx := range rr {
			if (rr[idx].Prerelease != nil && *rr[idx].Prerelease) ||
				(rr[idx].Draft != nil && *rr[idx].Draft) {
				continue
			}

			tagName := rr[idx].GetTagName()
			relName := rr[idx].GetName()
			version := relName

			log.Debug(fmt.Sprintf(
				"[%s] Analyzing release %s - %s...",
				atom.Name, relName, tagName))

			if !atom.GithubIgnoreTags() {
				validTags[tagName], present = tagsMap[tagName]
				if !present {
					// OMG! There are releases where the tag name is not equal to the real tag name.
					// For example: cbindgen has tag v0.29.0 and release 0.29.0 but rr[idx].GetTagName() returns 0.29.0
					if relName != "" && !strings.HasPrefix(relName, "v") {
						// Try to check if exists tag with prefix v
						tagName = "v" + relName
						validTags[tagName], present = tagsMap[tagName]
					}
					if !present {
						if log.Config.GetGeneral().Debug {
							log.Debug(fmt.Sprintf(
								"[%s] Release %s without tag. Skipped. Try to increase pages.",
								atom.Name, relName))
						}
						continue
					}
				}
				if version == "" {
					// POST: The release is without a valid name. Using tag name as fallback.
					version = tagName
				}
			} // POST Using release name as valid version

			if r.MatchString(version) {
				version = version[1:]
			}

			if matchRegex != nil && (!matchRegex.MatchString(version)) {
				log.Debug(fmt.Sprintf(
					"[%s] Release %s doesn't match with regex. Ignore it.",
					atom.Name, version))
				continue
			}

			versions = append(versions, version)

			validReleases[version] = rr[idx]
		}

		ans["releases"] = validReleases

	}

	// Retrive metadata of github repository
	repository, _, err := g.Client.Repositories.Get(
		ctx, atom.Github.User, atom.Github.Repo,
	)
	if err == nil {
		ans["repository"] = repository
		ans["desc"] = strings.ReplaceAll(repository.GetDescription(), "`", "")
		if repository.GetHomepage() != "" {
			ans["homepage"] = strings.ReplaceAll(repository.GetHomepage(), "`", "")
		} else {
			ans["homepage"] = repository.GetHTMLURL()
		}
		ans["github_fullname"] = repository.GetFullName()
		license := repository.GetLicense()
		if license != nil && license.SPDXID != nil {
			ans["license"] = *license.SPDXID
		}
	}

	ans["versions"] = versions
	ans["tags"] = validTags
	ans["github_user"] = atom.Github.User
	ans["github_repo"] = atom.Github.Repo
	ans["git_repo"] = fmt.Sprintf("https://github.com/%s/%s.git",
		atom.Github.User, atom.Github.Repo)

	return &ans, nil
}
