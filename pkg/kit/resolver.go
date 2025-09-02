/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	log "github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	gentoo "github.com/geaaru/pkgs-checker/pkg/gentoo"
)

type RepoScanResolver struct {
	Config *specs.MarkDevkitConfig
	Logger *log.MarkDevkitLogger

	JsonSources        []string
	Sources            []specs.RepoScanSpec
	Constraints        []string
	MapConstraints     map[string]([]gentoo.GentooPackage)
	Map                map[string]([]specs.RepoScanAtom)
	IgnoreMissingDeps  bool
	ContinueWithError  bool
	DepsWithSlot       bool
	AllowEmptyKeywords bool
	DisabledUseFlags   []string
	DisabledKeywords   []string
}

type PortageResolverOpts struct {
	EnableUseFlags   []string
	DisabledUseFlags []string
	Conditions       []string
	IgnoreSlot       bool
}

func NewPortageResolverOpts() *PortageResolverOpts {
	return &PortageResolverOpts{
		EnableUseFlags:   []string{},
		DisabledUseFlags: []string{},
		Conditions:       []string{},
		IgnoreSlot:       false,
	}
}

func (o *PortageResolverOpts) IsAdmitUseFlag(u string) bool {
	ans := true
	if len(o.EnableUseFlags) > 0 {
		for _, ue := range o.EnableUseFlags {
			if ue == u {
				return true
			}
		}

		return false
	}

	if len(o.DisabledUseFlags) > 0 {
		for _, ud := range o.DisabledUseFlags {
			if ud == u {
				ans = false
				break
			}
		}
	}

	return ans
}

func NewRepoScanResolver(c *specs.MarkDevkitConfig) *RepoScanResolver {
	return &RepoScanResolver{
		Config:             c,
		Logger:             log.GetDefaultLogger(),
		JsonSources:        make([]string, 0),
		Sources:            make([]specs.RepoScanSpec, 0),
		Constraints:        make([]string, 0),
		MapConstraints:     make(map[string][]gentoo.GentooPackage, 0),
		Map:                make(map[string][]specs.RepoScanAtom, 0),
		IgnoreMissingDeps:  false,
		DepsWithSlot:       true,
		ContinueWithError:  true,
		AllowEmptyKeywords: false,
	}
}

func (r *RepoScanResolver) SetContinueWithError(v bool) { r.ContinueWithError = v }
func (r *RepoScanResolver) GetContinueWithError() bool  { return r.ContinueWithError }

func (r *RepoScanResolver) SetIgnoreMissingDeps(v bool)    { r.IgnoreMissingDeps = v }
func (r *RepoScanResolver) IsIgnoreMissingDeps() bool      { return r.IgnoreMissingDeps }
func (r *RepoScanResolver) SetDepsWithSlot(v bool)         { r.DepsWithSlot = v }
func (r *RepoScanResolver) GetDepsWithSlot() bool          { return r.DepsWithSlot }
func (r *RepoScanResolver) SetDisabledUseFlags(u []string) { r.DisabledUseFlags = u }
func (r *RepoScanResolver) GetDisabledUseFlags() []string  { return r.DisabledUseFlags }
func (r *RepoScanResolver) SetDisabledKeywords(k []string) { r.DisabledKeywords = k }
func (r *RepoScanResolver) GetDisabledKeywords() []string  { return r.DisabledKeywords }
func (r *RepoScanResolver) SetAllowEmptyKeywords(v bool)   { r.AllowEmptyKeywords = v }
func (r *RepoScanResolver) GetAllowEmptyKeywords() bool    { return r.AllowEmptyKeywords }
func (r *RepoScanResolver) IsDisableUseFlag(u string) bool {
	ans := false

	if len(r.DisabledUseFlags) > 0 {
		for _, useFlag := range r.DisabledUseFlags {
			if useFlag == u {
				ans = true
				break
			}
		}
	}

	return ans
}

func (r *RepoScanResolver) LoadJson(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	decoder := json.NewDecoder(fd)

	for {
		var spec specs.RepoScanSpec
		if err := decoder.Decode(&spec); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		spec.File = path
		r.Sources = append(r.Sources, spec)
	}

	return nil
}

func (r *RepoScanResolver) LoadRawJson(raw, file string) error {
	var spec specs.RepoScanSpec
	err := json.Unmarshal([]byte(raw), &spec)
	if err != nil {
		return err
	}
	spec.File = file

	r.Sources = append(r.Sources, spec)

	return nil
}

func (r *RepoScanResolver) LoadJsonFiles(verbose bool) error {
	for _, file := range r.JsonSources {
		if verbose {
			r.Logger.InfoC(fmt.Sprintf(":brain:Loading reposcan file %s...", file))
		}
		err := r.LoadJson(file)
		if err != nil {
			return err
		}
	}

	// Create packages map
	return nil
}

func (r *RepoScanResolver) BuildMap() error {
	//fmt.Println("Build MAP ")
	for idx, _ := range r.Sources {

		for pkg, atom := range r.Sources[idx].Atoms {

			p := atom.CatPkg

			if atom.Status != "" {
				r.Logger.Warning(fmt.Sprintf(
					":warn Skipping pkg %s with wrong status.", pkg))
				// Skip broken atoms
				continue
			}

			if val, ok := r.Map[p]; ok {

				atomref := r.Sources[idx].Atoms[pkg]
				// POST: entry found
				r.Map[p] = append(val, atomref)

			} else {
				atomref := r.Sources[idx].Atoms[pkg]
				// POST: no entry available.
				r.Map[p] = []specs.RepoScanAtom{atomref}
			}
		}
	}

	// Build contraints Map
	if len(r.Constraints) > 0 {
		for _, c := range r.Constraints {
			gp, err := gentoo.ParsePackageStr(c)
			if err != nil {
				return err
			}

			if val, ok := r.MapConstraints[gp.GetPackageName()]; ok {
				r.MapConstraints[gp.GetPackageName()] = append(val, *gp)
			} else {
				r.MapConstraints[gp.GetPackageName()] = []gentoo.GentooPackage{*gp}
			}

		}
	}

	return nil
}

func (r *RepoScanResolver) GetMap() map[string]([]specs.RepoScanAtom) {
	return r.Map
}

func (r *RepoScanResolver) IsPresentPackage(pkg string) bool {
	_, ok := r.Map[pkg]
	return ok
}

func (r *RepoScanResolver) GetPackageVersions(pkg string) ([]specs.RepoScanAtom, bool) {
	ans, ok := r.Map[pkg]
	return ans, ok
}

func (r *RepoScanResolver) AddPackageAtom(pkg string, atom *specs.RepoScanAtom) {
	atoms, ok := r.Map[pkg]
	if ok {
		r.Map[pkg] = append(atoms, *atom)
	} else {
		r.Map[pkg] = []specs.RepoScanAtom{*atom}
	}
}

func (r *RepoScanResolver) GetValidPackages(pkg string, opts *PortageResolverOpts) ([]*specs.RepoScanAtom, error) {
	mAtoms := make(map[string]*specs.RepoScanAtom, 0)
	ans := []*specs.RepoScanAtom{}
	pkgs := []gentoo.GentooPackage{}

	gp, err := gentoo.ParsePackageStr(pkg)
	if err != nil {
		return nil, err
	}
	// Reset slot if not in input
	if strings.Index(pkg, ":") < 0 {
		gp.Slot = ""
	}
	// Ignore sub slot
	if strings.Contains(gp.Slot, "/") {
		gp.Slot = gp.Slot[0:strings.Index(gp.Slot, "/")]
	}

	atoms, ok := r.Map[gp.GetPackageName()]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Package (%s) %s not found in map.",
			gp.Condition.String(), gp.GetPackageName()))
	}

	if len(atoms) > 0 {
		for idx, atom := range atoms {
			p, err := atom.ToGentooPackage()
			if err != nil {
				// If the version is not supported, skip the version
				r.Logger.Warning(fmt.Sprintf(
					"[%s-%s::%s] Error on generate Gentoo package: %s. Package skipped.",
					atom.Atom, atom.Revision, atom.Kit, err.Error()))
				continue
			}

			if gp.Repository != "" {
				// Exclude package from different kit
				if gp.Repository != atom.Kit {
					r.Logger.Debug(fmt.Sprintf(
						"[%s-%s::%s] Skipping atom with kit %s != %s.",
						atom.Atom, atom.Revision, atom.Kit,
						atom.Kit, gp.Repository))
					continue
				}
			}

			// TODO: check of handle this in a better way
			valid, err := r.KeywordsIsAdmit(&atom, p)
			if err != nil {
				r.Logger.DebugC(fmt.Sprintf(
					"[%s-%s::%s] Check %s/%s:%s@%s: Invalid keyword.",
					atom.Atom, atom.Revision, atom.Kit,
					p.Category, p.GetPF(), p.Slot, p.Repository))
			}

			if valid {
				valid, err = r.PackageIsAdmit(gp, p, opts)
				if err != nil {
					r.Logger.Warning(fmt.Sprintf(
						"[%s/%s-%s] %s/%s:%s@%s: Package invalid: %s.",
						atom.Category, atom.Package, atom.Revision,
						p.Category, p.GetPF(), p.Slot, p.Repository, err.Error()))
					continue
				}
			} else {
				r.Logger.DebugC(fmt.Sprintf(
					"[%s/%s-%s] %s/%s:%s@%s: Not valid.",
					atom.Category, atom.Package, atom.Revision,
					p.Category, p.GetPF(), p.Slot, p.Repository))
			}

			r.Logger.DebugC(fmt.Sprintf(
				"[%s/%s:%s] Check (%s) %s/%s:%s@%s: admitted - %v",
				gp.Category, gp.GetPF(), gp.Slot, pkg,
				p.Category, p.GetPF(), p.Slot, p.Repository, valid))

			if valid {
				mAtoms[p.GetPVR()] = &atoms[idx]
				pkgs = append(pkgs, *p)
			}
		}

	}

	if len(pkgs) > 0 {
		sort.Sort(gentoo.GentooPackageSorter(pkgs))

		for idx := range pkgs {
			atom, ok := mAtoms[pkgs[idx].GetPVR()]
			if !ok {
				return nil, fmt.Errorf("unexpected error on retrieve atom for package %s!",
					pkgs[idx].GetPVR())
			}
			ans = append(ans, atom)
		}
	}

	return ans, nil
}

func (r *RepoScanResolver) GetLastPackage(pkg string, opts *PortageResolverOpts) (*specs.RepoScanAtom, error) {
	var ans *specs.RepoScanAtom

	atoms, err := r.GetValidPackages(pkg, opts)
	if err != nil {
		return ans, err
	}

	if len(atoms) > 0 {
		// POST: the atoms are sorted. I need the last.
		ans = atoms[len(atoms)-1]
	} else {
		if len(opts.Conditions) > 0 {
			return nil, fmt.Errorf("No packages found matching %s with defined conditions.", pkg)
		} else {
			return nil, fmt.Errorf("No packages found matching %s.", pkg)
		}
	}

	r.Logger.DebugC(
		fmt.Sprintf("[%s] Using package %s:%s",
			pkg, ans.Atom, ans.GetMetadataValue("SLOT")))

	return ans, nil
}

func (r *RepoScanResolver) PackageIsAdmit(target, atom *gentoo.GentooPackage,
	opts *PortageResolverOpts) (bool, error) {
	valid, err := target.Admit(atom)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, nil
	}

	// Check if atom is admitted by constraints
	if len(r.Constraints) > 0 {

		constraints, ok := r.MapConstraints[target.GetPackageName()]
		if ok {
			admitted := false

			for _, c := range constraints {

				admitted, err = c.Admit(atom)
				if err != nil {
					return false, err
				}
				if admitted {
					break
				}
			}

			if !admitted {
				r.Logger.DebugC(fmt.Sprintf("[%s] Package not admitted by constraints",
					atom.GetPF()))
			}
			valid = admitted
		} else {
			r.Logger.DebugC(fmt.Sprintf("[%s] No constraints found.",
				atom.GetPF()))
		}

	}

	if len(opts.Conditions) > 0 && valid {
		for _, cond := range opts.Conditions {
			p, err := gentoo.ParsePackageStr(cond)
			if err != nil {
				return valid, fmt.Errorf("Package %s has invalid condition %s: %s",
					atom.GetPackageName(), cond, err.Error())
			}

			if opts.IgnoreSlot {
				// Ensure that the slot are always equals.
				p.Slot = atom.Slot
			}
			ok, err := p.Admit(atom)
			if err != nil {
				return valid, fmt.Errorf("Package %s fail on check condition %s: %s",
					atom.GetPackageName(), cond, err.Error())
			}
			if !ok {
				valid = false
				r.Logger.DebugC(fmt.Sprintf("[%s] Package not admitted by condition %s",
					atom.GetPF(), cond))
				break
			}
		}

	}

	return valid, nil
}

func (r *RepoScanResolver) KeywordsIsAdmit(atom *specs.RepoScanAtom, p *gentoo.GentooPackage) (bool, error) {
	ans := true

	keywords := atom.GetMetadataValue("KEYWORDS")
	if keywords == "" && !r.AllowEmptyKeywords {
		r.Logger.DebugC(fmt.Sprintf(
			"[%s] Skip version without keywords %s or disabled.", atom.Atom, p.GetPF()))
		return false, nil
	}

	r.Logger.DebugC(fmt.Sprintf("[%s] Found KEYWORDS %s", atom.Atom, keywords))

	// On Portage it's possible a condition like this:
	// KEYWORDS="-* ~amd64"
	// This means that all keywords are disabled excluded ~amd64.
	// So, if i disabled -* i can accept the keywords with ~amd64.

	if len(r.DisabledKeywords) > 0 {
		ak := strings.Split(keywords, " ")
		for _, k := range ak {
			// We need to check all keywords every time
			for _, d := range r.DisabledKeywords {
				if d == k {
					ans = false
					break
				} else if !ans {
					ans = true
				}
			}
		}

		if !ans {
			r.Logger.DebugC(fmt.Sprintf(
				"[%s] Version %s disabled for keywords %s", atom.Atom, p.GetPF(), keywords))
		}
	}

	r.Logger.DebugC(fmt.Sprintf("[%s] KEYWORDS admit %v", atom.Atom, ans))

	return ans, nil
}
