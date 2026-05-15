/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	client "code.forgejo.org/f3/gof3/v3/forges/forgejo/sdk"
)

type ForgejoGenerator struct {
	Client *client.Client
	Host   string
}

func NewForgejoGenerator(opts map[string]string) *ForgejoGenerator {
	var err error
	log := logger.GetDefaultLogger()

	ans := &ForgejoGenerator{
		Client: nil,
	}

	if host, present := opts["host"]; present {

		opts := []client.ClientOption{}
		// Parse host string to retrieve host without protocol string.
		uri, _ := url.Parse(host)
		// Retrieve token
		remote, remotePresent := log.Config.GetAuthentication().GetRemote(uri.Host)

		if remotePresent && remote.Token != "" {
			opts = append(opts, client.SetToken(remote.Token))
		}

		ans.Client, err = client.NewClient(host, opts...)
		if err != nil {
			log.Error(fmt.Sprintf("error on setup forgejo client for host %s: %s",
				host, err.Error()))
		}
		ans.Host = host

	} else {
		log.Error(fmt.Sprintf("missed host field for forgejo generator"))
	}

	return ans
}

func (g *ForgejoGenerator) GetType() string {
	return specs.GeneratorBuiltinForgejo
}

func (g *ForgejoGenerator) SetClient(c *client.Client) { g.Client = c }

func (g *ForgejoGenerator) GetAssets(atom *specs.AutogenAtom,
	release *client.Release,
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

			for idx := range release.Attachments {
				if r.MatchString(release.Attachments[idx].Name) {
					assetFound = true
					ans = append(ans, &specs.AutogenArtefact{
						SrcUri: []string{release.Attachments[idx].DownloadURL},
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

func (g *ForgejoGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {
	log := logger.GetDefaultLogger()

	values := *mapref

	var tag *client.Tag
	var release *client.Release
	var err error
	var sha string

	originalVersion, _ := values["original_version"].(string)
	pv, _ := values["pv"].(string)

	// Set release metadata if present
	if _, present := values["releases"]; present {

		releases, _ := values["releases"].(map[string]*client.Release)
		release = releases[originalVersion]
		values["release"] = release

		tags, _ := values["tags"].(map[string]*client.Tag)
		tag = tags[release.TagName]
		if tag == nil {
			// Try to search tag name with v. Really i hate this.
			tag = tags["v"+release.TagName]
		}
		values["tag"] = tag

	} else {
		// Set only the tag
		tags, _ := values["tags"].(map[string]*client.Tag)
		tag = tags[originalVersion]
		values["tag"] = tag

	}

	if tag != nil {
		sha = tag.Commit.SHA
	} else if atom.ForgejoIgnoreTags() && release != nil {
		sha = release.Target
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
		tarballVersion := version
		if pv != "" {
			tarballVersion = pv
		}

		// Using sha at the end to correctly catch issues
		// with retag done on upstream repo
		if sha != "" && len(sha) > 7 {
			tarballName = fmt.Sprintf("%s-%s-%s.tar.gz", atom.Name, tarballVersion,
				sha[0:7])
		} else {
			tarballName = fmt.Sprintf("%s-%s.tar.gz", atom.Name, tarballVersion)
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
	} else if release != nil && release.TarURL != "" {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{release.TarURL},
			Name:   tarballName,
		})

	} else if tag != nil {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{tag.TarballURL},
			Name:   tarballName,
		})
	}

	values["artefacts"] = artefacts

	if sha != "" && len(sha) > 7 {
		values["pkg_basedir"] = fmt.Sprintf("%s-%s-%s",
			atom.Forgejo.User, atom.Forgejo.Repo, sha[0:7],
		)
	} else {
		values["pkg_basedir"] = fmt.Sprintf("%s-%s-%s",
			atom.Forgejo.User, atom.Forgejo.Repo, originalVersion,
		)
	}

	return nil
}

func (g *ForgejoGenerator) getTags(atom *specs.AutogenAtom) (
	[]*client.Tag, error) {

	tt := []*client.Tag{}
	log := logger.GetDefaultLogger()

	tOpts := client.ListRepoTagsOptions{
		ListOptions: client.ListOptions{
			Page: 1,
		},
	}
	if atom.Forgejo.PerPage != nil {
		tOpts.ListOptions.PageSize = *atom.Forgejo.PerPage
	}

	if atom.Forgejo.NumPages != nil {

		for page := 1; page < *atom.Forgejo.NumPages; page++ {
			tags, resp, err := g.Client.ListRepoTags(
				atom.Forgejo.User, atom.Forgejo.Repo,
				tOpts,
			)
			if err != nil {
				return tt, err
			}

			tt = append(tt, tags...)

			if log.Config.GetGeneral().Debug {
				for _, t := range tags {
					log.Debug(fmt.Sprintf(
						"[%s] Found tag %s at page %d.",
						atom.Name, strings.ReplaceAll(t.Name, "\n", ""), page))
				}
			}

			tOpts.Page = resp.NextPage
			if tOpts.Page > resp.LastPage {
				break
			}
		}

	} else {
		// POST: Read only one page.
		tags, _, err := g.Client.ListRepoTags(
			atom.Forgejo.User, atom.Forgejo.Repo, tOpts,
		)
		if err != nil {
			return tt, err
		}

		if log.Config.GetGeneral().Debug {
			for _, t := range tags {
				log.Debug(fmt.Sprintf(
					"[%s] Found tag %s.",
					atom.Name, strings.ReplaceAll(t.Name, "\n", "")))
			}
		}

		tt = tags
	}

	return tt, nil
}

func (g *ForgejoGenerator) getReleases(atom *specs.AutogenAtom) (
	[]*client.Release, error) {

	rr := []*client.Release{}
	log := logger.GetDefaultLogger()

	preRelease := false
	rOpts := client.ListReleasesOptions{
		ListOptions: client.ListOptions{
			Page: 1,
		},
		IsDraft:      &preRelease,
		IsPreRelease: &preRelease,
	}
	if atom.Forgejo.PerPage != nil {
		rOpts.ListOptions.PageSize = *atom.Forgejo.PerPage
	}

	if atom.Forgejo.NumPages != nil {

		for page := 1; page < *atom.Forgejo.NumPages; page++ {
			releases, resp, err := g.Client.ListReleases(
				atom.Forgejo.User, atom.Forgejo.Repo, rOpts,
			)
			if err != nil {
				return rr, err
			}

			rr = append(rr, releases...)

			if log.Config.GetGeneral().Debug {
				for _, r := range releases {
					log.Debug(fmt.Sprintf(
						"[%s] Found release %s at page %d.",
						atom.Name, strings.ReplaceAll(r.Title, "\n", ""), page))
				}
			}

			rOpts.ListOptions.Page = resp.NextPage
			if rOpts.ListOptions.Page > resp.LastPage {
				break
			}
		}

	} else {
		// POST: Read only one page.
		releases, _, err := g.Client.ListReleases(
			atom.Forgejo.User, atom.Forgejo.Repo, rOpts,
		)
		if err != nil {
			return rr, err
		}

		if log.Config.GetGeneral().Debug {
			for _, r := range releases {
				log.Debug(fmt.Sprintf(
					"[%s] Found release %s.",
					atom.Name, strings.ReplaceAll(r.Title, "\n", "")))
			}
		}

		rr = releases
	}

	return rr, nil
}

func (g *ForgejoGenerator) Process(atom *specs.AutogenAtom) (
	*map[string]interface{}, error) {

	ans := make(map[string]interface{}, 0)
	log := logger.GetDefaultLogger()
	var matchRegex *regexp.Regexp

	if g.Client == nil {
		return nil, fmt.Errorf("forgejo client not correctly initialized for atom %s",
			atom.Name)
	}

	// Use atom.Name as default value for forgejo repo and user if not defined.
	if atom.Forgejo.Repo == "" {
		atom.Forgejo.Repo = atom.Name
	}

	if atom.Forgejo.User == "" {
		atom.Forgejo.User = atom.Name
	}

	// Using release as default values for forgejo if not defined.
	if atom.Forgejo.Query == "" {
		atom.Forgejo.Query = "releases"
	}

	if atom.Forgejo.Query != "releases" && atom.Forgejo.Query != "tags" {
		return nil, fmt.Errorf("forgejo query with invalid query for atom %s",
			atom.Name)
	}

	validTags := make(map[string]*client.Tag, 0)
	versions := []string{}
	r := regexp.MustCompile("^v[0-9].*")

	if atom.Forgejo.Match != "" {
		matchRegex = regexp.MustCompile(atom.Forgejo.Match)
		if matchRegex == nil {
			return nil, fmt.Errorf("invalid regex match string for atom %s",
				atom.Name)
		}
	}

	if atom.Forgejo.Query == "tags" {

		tt, err := g.getTags(atom)
		if err != nil {
			return nil, err
		}

		for idx := range tt {
			version := tt[idx].Name

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

		rr, err := g.getReleases(atom)
		if err != nil {
			return nil, err
		}

		tagsMap := make(map[string]*client.Tag, 0)

		if !atom.ForgejoIgnoreTags() {
			tags, err := g.getTags(atom)
			if err != nil {
				return nil, err
			}

			for i := range tags {
				tagsMap[tags[i].Name] = tags[i]
			}
		}

		validReleases := make(map[string]*client.Release, 0)
		var present bool

		for idx := range rr {
			if rr[idx].IsPrerelease || rr[idx].IsDraft {
				continue
			}

			tagName := rr[idx].TagName
			relName := rr[idx].Title
			version := relName

			log.Debug(fmt.Sprintf(
				"[%s] Analyzing release %s - %s...", atom.Name, relName, tagName))

			if !atom.ForgejoIgnoreTags() {
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

	// Retrive metadata of forgejo repository
	repository, _, err := g.Client.GetRepo(
		atom.Forgejo.User, atom.Forgejo.Repo,
	)
	if err == nil {
		ans["repository"] = repository
		ans["desc"] = strings.ReplaceAll(repository.Description, "`", "")
		if repository.Website != "" {
			ans["homepage"] = strings.ReplaceAll(repository.Website, "`", "")
		} else {
			ans["homepage"] = repository.HTMLURL
		}
		ans["forgejo_fullname"] = repository.FullName
		ans["git_repo"] = repository.CloneURL
	} else {
		ans["git_repo"] = fmt.Sprintf("%s/%s/%s.git",
			g.Host, atom.Forgejo.User, atom.Forgejo.Repo)
	}

	ans["versions"] = versions
	ans["tags"] = validTags
	ans["forgejo_user"] = atom.Forgejo.User
	ans["forgejo_repo"] = atom.Forgejo.Repo

	return &ans, nil
}
