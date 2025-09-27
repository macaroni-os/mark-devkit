/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
)

type PypiGenerator struct {
	RestGuard *guard.RestGuard
}

func NewPypiGenerator() *PypiGenerator {
	log := logger.GetDefaultLogger()
	rg, _ := guard.NewRestGuard(log.Config.GetRest())
	// Overide the default check redirect
	rg.Client.CheckRedirect = kit.CheckRedirect
	return &PypiGenerator{
		RestGuard: rg,
	}
}

func (g *PypiGenerator) GetType() string {
	return specs.GeneratorBuiltinPypi
}

func (g *PypiGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	var err error
	values := *mapref

	originalVersion, _ := values["original_version"].(string)
	pypiMeta, _ := values["pypi_meta"].(*specs.PypiMetadata)

	delete(values, "versions")
	delete(values, "pypi_meta")

	artefacts := []*specs.AutogenArtefact{}

	err = g.processDependencies(atom, version, mapref, pypiMeta.GetInfo())
	if err != nil {
		return err
	}

	if atom.HasAssets() {

		pypiFiles := pypiMeta.GetReleaseFiles(originalVersion, "")
		for _, asset := range atom.Assets {

			name, err := helpers.RenderContentWithTemplates(
				asset.Name,
				"", "", "asset.name", values, []string{},
			)
			if err != nil {
				return err
			}

			r := regexp.MustCompile(asset.Matcher)
			if r == nil {
				return fmt.Errorf("[%s] invalid regex on asset %s", atom.Name, asset.Name)
			}

			assetFound := false

			for idx := range pypiFiles {
				if r.MatchString(pypiFiles[idx].Filename) {
					assetFound = true
					artefacts = append(artefacts, &specs.AutogenArtefact{
						SrcUri: []string{pypiFiles[idx].Url},
						Use:    asset.Use,
						Name:   name,
						Hashes: pypiFiles[idx].Digests,
					})
					break
				}
			}

			if !assetFound {
				return fmt.Errorf("[%s] no asset found for matcher %s", atom.Name, asset.Matcher)
			}

		}

	} else {
		pypiFiles := pypiMeta.GetReleaseFiles(originalVersion, "sdist")
		if len(pypiFiles) == 0 {
			return fmt.Errorf("[%s] no sdist files found.", atom.Name)
		}

		tarballName := atom.Tarball
		if tarballName == "" {
			tarballName = pypiFiles[0].Filename
		} else {
			tarballName, err = helpers.RenderContentWithTemplates(
				tarballName,
				"", "", "artefact.tarball", values, []string{},
			)
			if err != nil {
				return err
			}
		}

		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{pypiFiles[0].Url},
			Name:   tarballName,
			Hashes: pypiFiles[0].Digests,
		})

	}

	values["artefacts"] = artefacts
	values["pypi_version"] = originalVersion

	return nil
}

func (g *PypiGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {

	ans := make(map[string]interface{}, 0)

	pypiName := atom.Name
	if atom.Python.PypiName != "" {
		pypiName = atom.Python.PypiName
	}

	pypiJsonUrl := fmt.Sprintf("https://pypi.org/pypi/%s/json", pypiName)

	uri, err := url.Parse(pypiJsonUrl)
	if err != nil {
		return nil, err
	}

	ssl := true
	node := guard_specs.NewRestNode(uri.Host,
		uri.Host+path.Dir(uri.Path), ssl)
	resource := path.Base(uri.Path)

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

	data, err := io.ReadAll(t.Response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal json data
	pypiMeta := specs.NewPypiMetadata()
	if err = json.Unmarshal(data, pypiMeta); err != nil {
		return nil, err
	}

	info := pypiMeta.GetInfo()
	// Improve readable on debug
	info.Description = ""

	ans["pypi_meta"] = pypiMeta
	ans["pypi_info"] = info
	if atom.Python.PypiName != "" {
		ans["pypi_name"] = atom.Python.PypiName
	} else {
		ans["pypi_name"] = info.Name
	}
	// We need to ignore char ` that is expanded to bash command.
	// The code injection over description must be blocked for multiple reasons.
	ans["desc"] = strings.ReplaceAll(info.Summary, "`", "'")
	ans["license"] = info.License
	if atom.Python.PythonCompat != "" {
		ans["python_compat"] = atom.Python.PythonCompat
	} else {
		ans["python_compat"] = "python3+"
	}

	if atom.Python.PythonRequiresIgnore != "" {
		ans["versions"] = pypiMeta.GetVersions("")
	} else {
		ans["versions"] = pypiMeta.GetVersions(atom.Python.PythonCompat)
	}

	if homepage, present := info.ProjectUrls["Homepage"]; present {
		ans["homepage"] = homepage
	} else if homepage, present := info.ProjectUrls["Source"]; present {
		ans["homepage"] = homepage
	}

	return &ans, nil
}

func (g *PypiGenerator) processDependencies(atom *specs.AutogenAtom,
	version string, mapref *map[string]interface{}, info *specs.PypiPackageInfo) error {
	log := logger.GetDefaultLogger()

	var err error
	values := *mapref
	pydepsRdepend := []*specs.PythonDep{}
	pydepsBdepend := []*specs.PythonDep{}
	pydepsPdepend := []*specs.PythonDep{}
	pydepsDepend := []*specs.PythonDep{}

	if len(atom.Python.Pydeps) > 0 {
		for k, deps := range atom.Python.Pydeps {

			mods := []string{}
			use := ""
			target := "rdepend"
			labels := strings.Split(k, ":")
			if len(labels) < 2 {
				// TODO: add warning
				continue
			}

			if labels[0] != "py" && labels[0] != "use" {
				// TODO: add warning
				continue
			}

			if len(labels) > 2 {

				switch labels[2] {
				case "build":
					target = "bdepend"
				case "post":
					target = "pdepend"
				case "tool":
					target = "depend"
				case "runtime":
					target = "rdepend"
				case "both":
					target = "both"
				default:
					// TODO: add warning for invalid label
					continue
				}
			}

			if labels[0] == "py" {
				mods = strings.Split(labels[1], ",")
			} else {
				// POST: labels[0] == "use"
				use = labels[1]
			}

			for _, depstr := range deps {

				pydep := &specs.PythonDep{
					Use:        use,
					PyVersions: mods,
				}
				words := strings.Split(depstr, " ")
				switch len(words) {
				case 1:
					pydep.Dependency = &gentoo.GentooPackage{
						Category: "dev-python",
						Name:     words[0],
					}
				case 2:
					continue
				case 3:
					gpkg, err := helpers.DecodeCondition(words[1]+words[2],
						"dev-python", words[0])
					if err != nil {
						return err
					}
					pydep.Dependency = gpkg
				default:
					continue
				}

				switch target {
				case "rdepend":
					pydepsRdepend = append(pydepsRdepend, pydep)
				case "bdepend":
					pydepsBdepend = append(pydepsBdepend, pydep)
				case "pdepend":
					pydepsPdepend = append(pydepsPdepend, pydep)
				case "both":
					pydepsRdepend = append(pydepsRdepend, pydep)
					pydepsDepend = append(pydepsDepend, pydep)
				default:
					pydepsDepend = append(pydepsDepend, pydep)
				}

			}

		}

	} else if len(info.RequiresDist) > 0 {

		// Prepare regex filters
		filters := []*regexp.Regexp{}

		if len(atom.Python.DepsIgnore) > 0 {
			for _, d := range atom.Python.DepsIgnore {
				rr := regexp.MustCompile(d)
				if rr != nil {
					filters = append(filters, rr)
				} else {
					// TODO: add warning
				}
			}

		}

		// Try to use requires from pypi JSON

		for _, str := range info.RequiresDist {

			if len(filters) > 0 {
				toSkip := false
				for _, r := range filters {
					if r.MatchString(str) {
						toSkip = true
						if log.Config.GetGeneral().Debug {
							log.Debug(fmt.Sprintf(
								"[%s] Skip depend %s filtered.",
								atom.Name, str))
						}
						break
					}
				}
				if toSkip {
					continue
				} else {
					log.Debug(fmt.Sprintf(
						"[%s] depend %s is valid.",
						atom.Name, str))
				}

			}

			// Ignore extra note
			rdWords := strings.Split(str, ";")

			pnameidx := len(rdWords[0])

			if strings.Index(rdWords[0], "<") > 0 {
				pnameidx = strings.Index(rdWords[0], "<")
			} else if strings.Index(rdWords[0], ">") > 0 && strings.Index(rdWords[0], ">") < pnameidx {
				pnameidx = strings.Index(rdWords[0], ">")
			} else if strings.Index(rdWords[0], "!=") > 0 && strings.Index(rdWords[0], "!=") < pnameidx {
				pnameidx = strings.Index(rdWords[0], "!=")
			} else if strings.Index(rdWords[0], "==") > 0 && strings.Index(rdWords[0], "==") < pnameidx {
				pnameidx = strings.Index(rdWords[0], "==")
			}

			pname := rdWords[0][0:pnameidx]
			multicond := rdWords[0][pnameidx:len(rdWords[0])]

			conds := strings.Split(multicond, ",")
			for _, cond := range conds {
				pydep := &specs.PythonDep{
					Use:        "",
					PyVersions: []string{"all"},
				}

				// Workaround for v2.0a0
				cond = strings.ReplaceAll(cond, "a0", "")

				gpkg, err := helpers.DecodeCondition(cond,
					"dev-python", pname)
				if err != nil {
					return err
				}

				if gpkg.Version == "" {
					// TODO: add warning
					continue
				}
				pydep.Dependency = gpkg
				pydepsDepend = append(pydepsDepend, pydep)
			}

			// I dunno if deps are for build or runtime.
			// Set the deps on both.
			pydepsRdepend = pydepsDepend
		}
	}

	if len(pydepsRdepend) > 0 {
		values["py_rdepend"], err = g.stringifyDeps(pydepsRdepend)
		if err != nil {
			return err
		}
	}
	if len(pydepsBdepend) > 0 {
		values["py_bdepend"], err = g.stringifyDeps(pydepsBdepend)
		if err != nil {
			return err
		}
	}
	if len(pydepsPdepend) > 0 {
		values["py_pdepend"], err = g.stringifyDeps(pydepsPdepend)
		if err != nil {
			return err
		}
	}
	if len(pydepsDepend) > 0 {
		values["py_depend"], err = g.stringifyDeps(pydepsDepend)
	}

	return err
}

func (g *PypiGenerator) stringifyDeps(deps []*specs.PythonDep) (string, error) {
	ans := ""

	for _, dep := range deps {
		// TODO: Manage py_versions with specific release.
		pkgstr := ""

		if dep.Dependency.Condition == gentoo.PkgCondInvalid {
			pkgstr = dep.Dependency.GetPackageName() + "[${PYTHON_USEDEP}]"
		} else {
			pkgstr = fmt.Sprintf("%s%s-%s[${PYTHON_USEDEP}]",
				dep.Dependency.Condition.String(),
				dep.Dependency.GetPackageName(),
				dep.Dependency.GetPVR())
		}

		if dep.Use != "" {
			ans += fmt.Sprintf("\n%s? ( %s )", dep.Use, pkgstr)
		} else {
			ans += fmt.Sprintf("\n%s", pkgstr)
		}
	}

	return ans, nil
}
