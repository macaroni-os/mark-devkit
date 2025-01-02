/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"gopkg.in/yaml.v3"
)

func NewReposcanAnalysis(file string) (*ReposcanAnalysis, error) {
	ans := &ReposcanAnalysis{}

	content, err := os.ReadFile(file)
	if err != nil {
		return ans, err
	}

	err = yaml.Unmarshal(content, ans)
	return ans, err
}

func (ra *ReposcanAnalysis) Yaml() ([]byte, error) {
	return yaml.Marshal(ra)
}

func (ra *ReposcanAnalysis) Json() ([]byte, error) {
	return json.Marshal(ra)
}

func (ra *ReposcanAnalysis) GetKitsEclassDirs(cloneDir string) ([]string, error) {
	// Prepare eclass dir list
	eclassDirs := []string{}
	for _, source := range ra.Kits {
		eclassDir, err := filepath.Abs(filepath.Join(cloneDir, source.Name, "eclass"))
		if err != nil {
			return eclassDirs, err
		}
		if utils.Exists(eclassDir) {
			kitDir, _ := filepath.Abs(filepath.Join(cloneDir, source.Name))
			eclassDirs = append(eclassDirs, kitDir)
		}
	}

	return eclassDirs, nil
}

func (ra *ReposcanAnalysis) WriteYamlFile(file string) error {
	data, err := ra.Yaml()
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

func NewReposcanKit(name, url, branch, commit string) *ReposcanKit {
	return &ReposcanKit{
		Name:       name,
		Url:        url,
		Branch:     branch,
		CommitSha1: commit,
	}
}

func (r *ReposcanKit) GetPriority() int {
	if r.Priority != nil {
		return *r.Priority
	}
	return 1
}

func (r *RepoScanSpec) Yaml() (string, error) {
	data, err := yaml.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *RepoScanSpec) Json() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *RepoScanSpec) WriteJsonFile(f string) error {
	// TODO: Check if using writer from json marshal
	//       could be better
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func (r *RepoScanAtom) GetPackageName() string {
	return fmt.Sprintf("%s/%s", r.GetCategory(), r.Package)
}

func (r *RepoScanAtom) GetCategory() string {
	slot := "0"

	if r.HasMetadataKey("SLOT") {
		slot = r.GetMetadataValue("SLOT")
		// We ignore subslot atm.
		if strings.Contains(slot, "/") {
			slot = slot[0:strings.Index(slot, "/")]
		}

	}

	return SanitizeCategory(r.Category, slot)
}

func (r *RepoScanAtom) Yaml() (string, error) {
	data, err := yaml.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *RepoScanAtom) Json() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *RepoScanAtom) HasMetadataKey(k string) bool {
	_, ans := r.Metadata[k]
	return ans
}

func (r *RepoScanAtom) GetMetadataValue(k string) string {
	ans, _ := r.Metadata[k]
	return ans
}

func (r *RepoScanAtom) ToGentooPackage() (*gentoo.GentooPackage, error) {
	ans, err := gentoo.ParsePackageStr(r.Atom)
	if err != nil {
		return nil, err
	}

	// Retrieve license
	if l, ok := r.Metadata["LICENSE"]; ok {
		ans.License = l
	}

	if slot, ok := r.Metadata["SLOT"]; ok {
		// TOSEE: We ignore subslot atm.
		if strings.Contains(slot, "/") {
			slot = slot[0:strings.Index(slot, "/")]
		}
		ans.Slot = slot
	}

	ans.Repository = r.Kit

	return ans, nil
}

func (r *RepoScanAtom) AddRelations(pkgname string) {
	isPresent := false
	for idx := range r.Relations {
		if r.Relations[idx] == pkgname {
			isPresent = true
			break
		}
	}

	if !isPresent {
		r.Relations = append(r.Relations, pkgname)
	}
}

func (r *RepoScanAtom) AddRelationsByKind(kind, pkgname string) {
	isPresent := false
	list, kindPresent := r.RelationsByKind[kind]

	if kindPresent {
		for idx := range list {
			if list[idx] == pkgname {
				isPresent = true
				break
			}
		}
	} else {
		r.RelationsByKind[kind] = []string{}
	}

	if !isPresent {
		r.RelationsByKind[kind] = append(r.RelationsByKind[kind], pkgname)
	}
}

func (r *RepoScanAtom) GetRuntimeDeps() ([]gentoo.GentooPackage, error) {
	ans := []gentoo.GentooPackage{}

	if len(r.Relations) > 0 {
		if _, ok := r.RelationsByKind["RDEPEND"]; ok {

			deps, err := r.getDepends("RDEPEND")
			if err != nil {
				return ans, err
			}
			ans = append(ans, deps...)
		}
		// TODO: Check if it's needed add PDEPEND here
	}

	return ans, nil
}
func (r *RepoScanAtom) GetBuildtimeDeps() ([]gentoo.GentooPackage, error) {
	ans := []gentoo.GentooPackage{}

	if len(r.Relations) > 0 {
		if _, ok := r.RelationsByKind["DEPEND"]; ok {
			deps, err := r.getDepends("DEPEND")
			if err != nil {
				return ans, err
			}
			ans = append(ans, deps...)
		}

		if _, ok := r.RelationsByKind["BDEPEND"]; ok {
			deps, err := r.getDepends("BDEPEND")
			if err != nil {
				return ans, err
			}
			ans = append(ans, deps...)
		}
	}

	return ans, nil
}

func (r *RepoScanAtom) getDepends(depType string) ([]gentoo.GentooPackage, error) {
	ans := []gentoo.GentooPackage{}
	if _, ok := r.RelationsByKind[depType]; ok {

		for _, pkg := range r.RelationsByKind[depType] {
			gp, err := gentoo.ParsePackageStr(pkg)
			if err != nil {
				return ans, err
			}
			gp.Slot = ""
			ans = append(ans, *gp)
		}
	}

	return ans, nil
}
