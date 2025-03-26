/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"golang.org/x/net/html"
)

type DirlistingGenerator struct {
	RestGuard *guard.RestGuard
}

func NewDirlistingGenerator() *DirlistingGenerator {
	log := logger.GetDefaultLogger()
	rg, _ := guard.NewRestGuard(log.Config.GetRest())
	// Overide the default check redirect
	rg.Client.CheckRedirect = kit.CheckRedirect
	return &DirlistingGenerator{
		RestGuard: rg,
	}
}

func (g *DirlistingGenerator) GetType() string {
	return specs.GeneratorBuiltinDirListing
}

func (g *DirlistingGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {
	var err error

	values := *mapref
	originalVersion, _ := values["original_version"].(string)
	links, _ := values["links"].(map[string]string)
	link, _ := links[originalVersion]
	urlBase, _ := values["url"].(string)

	tarballName := atom.Tarball
	if tarballName == "" {
		// Using name from links
		tarballName = path.Base(link)
	} else {
		tarballName, err = helpers.RenderContentWithTemplates(
			tarballName,
			"", "", "artefact.tarball", values, []string{},
		)
		if err != nil {
			return err
		}
	}

	delete(values, "versions")
	delete(values, "links")

	artefacts := []*specs.AutogenArtefact{}

	if atom.HasAssets() {
		for _, asset := range atom.Assets {
			name, err := helpers.RenderContentWithTemplates(
				asset.Name,
				"", "", "asset.name", values, []string{},
			)
			if err != nil {
				return err
			}

			srcUri := ""
			if asset.Url != "" {
				// POST: We use the url value as urlBase
				srcUri, err = helpers.RenderContentWithTemplates(
					asset.Url,
					"", "", "asset.url", values, []string{},
				)
				if err != nil {
					return err
				}

			} else {
				// POST: we use urlbase as dir.url field.
				srcUri = urlBase
				if !strings.HasSuffix(urlBase, "/") {
					srcUri += "/"
				}

				if asset.Prefix != "" {
					prefix, err := helpers.RenderContentWithTemplates(
						asset.Prefix,
						"", "", "asset.prefix", values, []string{},
					)
					if err != nil {
						return err
					}
					srcUri += prefix
				}
				srcUri += name
			}

			artefacts = append(artefacts, &specs.AutogenArtefact{
				SrcUri: []string{srcUri},
				Use:    asset.Use,
				Name:   name,
			})
		}

	} else {
		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{link},
			Name:   tarballName,
		})
	}

	values["artefacts"] = artefacts

	return nil
}

func (g *DirlistingGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)

	if atom.Dir.Matcher == "" {
		return nil, fmt.Errorf("[%s] No matcher defined!", atom.Name)
	}
	if atom.Dir.Url == "" {
		return nil, fmt.Errorf("[%s] No url defined!", atom.Name)
	}

	var rexclude *regexp.Regexp = nil

	r := regexp.MustCompile(atom.Dir.Matcher)
	if r == nil {
		return nil, fmt.Errorf("[%s] invalid regex on matcher", atom.Name)
	}
	if atom.Dir.ExcludesMatcher != "" {
		rexclude = regexp.MustCompile(atom.Dir.ExcludesMatcher)
		if rexclude == nil {
			return nil, fmt.Errorf("[%s] invalid regex on exclude", atom.Name)
		}
	}

	uri, err := url.Parse(atom.Dir.Url)
	if err != nil {
		return nil, err
	}

	ssl := false

	if uri.Scheme == "ftp" {
		return nil, fmt.Errorf("Not yet implemented")
	}

	node := guard_specs.NewRestNode(uri.Host,
		uri.Host+path.Dir(uri.Path), ssl)

	resource := ""
	if !strings.HasSuffix(atom.Dir.Url, "/") {
		resource = path.Base(uri.Path)
	}

	service := guard_specs.NewRestService(uri.Host)
	service.Retries = 3
	service.AddNode(node)

	t := service.GetTicket()
	defer t.Rip()

	_, err = g.RestGuard.CreateRequest(t, "GET", "/"+resource)
	if err != nil {
		return nil, err
	}

	err = g.RestGuard.Do(t)
	if err != nil {
		if t.Response != nil {
			return nil, fmt.Errorf("%s - %s - %s", uri.Path, err.Error(), t.Response.Status)
		} else {
			return nil, fmt.Errorf("%s - %s", uri.Path, err.Error())
		}
	}

	if t.Response.Body == nil {
		return nil, fmt.Errorf("%s - Received invalid response body", uri.Path)
	}

	doc, err := html.Parse(t.Response.Body)
	if err != nil {
		return nil, err
	}

	links := make(map[string]string, 0)
	var versions []string
	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && r.MatchString(attr.Val) {
					if rexclude != nil {
						if rexclude.MatchString(attr.Val) {
							continue
						}
					}
					versions = append(versions, path.Base(attr.Val))

					if strings.HasPrefix(attr.Val, "https") || strings.HasPrefix(attr.Val, "http") {
						links[path.Base(attr.Val)] = attr.Val
					} else {
						if strings.HasSuffix(atom.Dir.Url, "/") {
							links[path.Base(attr.Val)] = atom.Dir.Url + attr.Val
						} else {
							links[path.Base(attr.Val)] = atom.Dir.Url + "/" + attr.Val
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}
	findLinks(doc)

	ans["url"] = atom.Dir.Url
	ans["versions"] = versions
	ans["links"] = links

	return &ans, nil
}
