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
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"golang.org/x/net/html"
)

type DirlistingGenerator struct {
	RestGuard *guard.RestGuard
}

func NewDirlistingGenerator() *DirlistingGenerator {
	rcfg := guard_specs.NewConfig()
	//rcfg.DisableCompression = true
	rg, _ := guard.NewRestGuard(rcfg)
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

			srcUri := urlBase
			if !strings.HasSuffix(urlBase, "/") {
				srcUri += "/"
			}
			srcUri += name

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

func (g *DirlistingGenerator) Process(atom *specs.AutogenAtom,
	def *specs.AutogenAtom) (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)

	dData := def.Clone()

	if atom.Dir != nil {
		if atom.Dir.Url != "" {
			dData.Dir.Url = atom.Dir.Url
		}
		if atom.Dir.Matcher != "" {
			dData.Dir.Matcher = atom.Dir.Matcher
		}
		if atom.Dir.ExcludesMatcher != "" {
			dData.Dir.ExcludesMatcher = atom.Dir.ExcludesMatcher
		}
	}

	if dData.Dir.Matcher == "" {
		return nil, fmt.Errorf("[%s] No matcher defined!", atom.Name)
	}
	if dData.Dir.Url == "" {
		return nil, fmt.Errorf("[%s] No url defined!", atom.Name)
	}

	var rexclude *regexp.Regexp = nil

	r := regexp.MustCompile(dData.Dir.Matcher)
	if r == nil {
		return nil, fmt.Errorf("[%s] invalid regex on matcher", atom.Name)
	}
	if dData.Dir.ExcludesMatcher != "" {
		rexclude = regexp.MustCompile(dData.Dir.ExcludesMatcher)
		if rexclude == nil {
			return nil, fmt.Errorf("[%s] invalid regex on exclude", atom.Name)
		}
	}

	uri, err := url.Parse(dData.Dir.Url)
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
	if !strings.HasSuffix(dData.Dir.Url, "/") {
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
						if strings.HasSuffix(dData.Dir.Url, "/") {
							links[path.Base(attr.Val)] = dData.Dir.Url + attr.Val
						} else {
							links[path.Base(attr.Val)] = dData.Dir.Url + "/" + attr.Val
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

	ans["url"] = dData.Dir.Url
	ans["versions"] = versions
	ans["links"] = links

	return &ans, nil
}
