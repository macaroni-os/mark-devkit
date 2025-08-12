/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package autogen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

func (a *AutogenBot) transformsVersions(atom *specs.AutogenAtom, versions []string) (*map[string]string, error) {
	ans := make(map[string]string, 0)
	var v string
	var r *regexp.Regexp

	for _, transform := range atom.Transforms {
		if transform.Kind == "regex" {
			r = regexp.MustCompile(transform.Match)
			if r == nil {
				return nil, fmt.Errorf("invalid regex string %s", transform.Match)
			}
		}

		for idx := range versions {
			if elabVer, ok := ans[versions[idx]]; ok {
				v = elabVer
			} else {
				v = versions[idx]
			}
			switch transform.Kind {
			case "string":
				v = strings.ReplaceAll(v, transform.Match, transform.Replace)
			case "regex":
				v = r.ReplaceAllString(v, transform.Replace)
			default:
				return nil, fmt.Errorf("unsupported kind of transform %s for atom %s",
					transform.Kind, atom.Name)
			}

			ans[versions[idx]] = v
		}
	}

	return &ans, nil
}

func (a *AutogenBot) sortVersions(atom *specs.AutogenAtom, versions []string) ([]string, error) {
	// In order to avoid issues with go-version on parse particular versions
	// I prefer to use GentooPackage that already support different versions and sorting.

	ans := []string{}
	pkgs := []gentoo.GentooPackage{}

	for idx := range versions {
		gp, err := gentoo.ParsePackageStr(fmt.Sprintf("%s/%s-%s",
			atom.Category,
			atom.Name,
			versions[idx],
		))
		if err != nil {
			return ans, err
		}
		if gp.Version == "" {
			a.Logger.Debug(fmt.Sprintf(
				":warning: [%s] Ignoring version '%s'", atom.Name, versions[idx]))
			continue
		}

		pkgs = append(pkgs, *gp)
	}

	sort.Sort(sort.Reverse(gentoo.GentooPackageSorter(pkgs)))

	for idx := range pkgs {
		ans = append(ans, pkgs[idx].GetPVR())
	}

	return ans, nil
}

func (a *AutogenBot) excludesVersions(atom *specs.AutogenAtom,
	versions []string) ([]string, error) {
	ans := []string{}
	excludes := []*regexp.Regexp{}

	// Prepare regexes
	for _, matcherRegex := range atom.Excludes {
		r := regexp.MustCompile(matcherRegex)
		if r == nil {
			return ans, fmt.Errorf("[%s] invalid regex %s on exclude",
				atom.Name, matcherRegex)
		}
		excludes = append(excludes, r)
	}

	for idx := range versions {
		toExclude := false
		for _, r := range excludes {
			if r.MatchString(versions[idx]) {
				toExclude = true
				break
			}
		}

		a.Logger.Debug(fmt.Sprintf(
			":eyes: [%s] Version %s to exclude: %v", atom.Name,
			versions[idx], toExclude))
		if toExclude {
			continue
		}
		ans = append(ans, versions[idx])
	}

	return ans, nil
}
