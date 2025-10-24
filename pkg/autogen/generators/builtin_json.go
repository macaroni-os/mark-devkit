/*
	Copyright © 2024-2025 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package generators

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"regexp"

	"github.com/macaroni-os/mark-devkit/pkg/helpers"
	"github.com/macaroni-os/mark-devkit/pkg/kit"
	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

type JsonGenerator struct {
	*BaseGenerator
	RestGuard   *guard.RestGuard
	MapServices map[string]*guard_specs.RestService
	RateLimit   string
}

func NewJsonGenerator(opts map[string]string) *JsonGenerator {
	log := logger.GetDefaultLogger()
	rg, _ := guard.NewRestGuard(log.Config.GetRest())

	// Overide the default check redirect
	rg.Client.CheckRedirect = kit.CheckRedirect
	ans := &JsonGenerator{
		BaseGenerator: NewBaseGenerator(opts),
		RestGuard:     rg,
	}

	// Set storage
	storage := *log.Config.GetStorage()
	mServicesI, ok := storage[specs.GeneratorBuiltinJson]
	if ok {
		ans.MapServices, _ = mServicesI.(map[string]*guard_specs.RestService)
	} else {
		// POST: storage is not initialized.
		ans.MapServices = make(map[string]*guard_specs.RestService, 0)
		storage[specs.GeneratorBuiltinJson] = ans.MapServices
	}

	if limit, present := opts[guard_specs.ServiceRateLimiter]; present {
		log.DebugC(fmt.Sprintf(
			":brain: Using rate limit %s...", limit))
		ans.RateLimit = limit
	}

	return ans
}

func (g *JsonGenerator) GetRestGuardService(service string) *guard_specs.RestService {
	var ans *guard_specs.RestService = nil
	if s, present := g.MapServices[service]; present {
		ans = s.Clone()
		// Ensure same rate limiter
		ans.RateLimiter = s.RateLimiter
	} else {
		ans = guard_specs.NewRestService(service)
		ans.Retries = 3
		ans.SetOption(guard_specs.ServiceRateLimiter, g.RateLimit)
		ans.SetRateLimiter()
	}

	return ans
}

func (g *JsonGenerator) GetType() string {
	return specs.GeneratorBuiltinJson
}

func (g *JsonGenerator) SetVersion(atom *specs.AutogenAtom, version string,
	mapref *map[string]interface{}) error {

	if atom.Json.FilterSrcUri != "" {
		values := *mapref
		artefacts, _ := values["artefacts"].([]*specs.AutogenArtefact)

		obj, _ := values["json_body"].(interface{})

		filterRendered, err := helpers.RenderContentWithTemplates(
			atom.Json.FilterSrcUri,
			"", "", "json.filter_srcuri", values, []string{},
		)
		if err != nil {
			return err
		}

		// Parse src uri filter
		srcUriFilter, err := jp.ParseString(filterRendered)
		if err != nil {
			return fmt.Errorf("error on parsing version_filter '%s': %s",
				filterRendered, err.Error())
		}

		res := srcUriFilter.Get(obj)
		srcUri := ""
		for _, v := range res {
			vstr, valid := v.(string)
			if !valid {
				return fmt.Errorf("error on convert src uri in string")
			}
			// Get the first
			srcUri = vstr
			break
		}

		uri, _ := url.Parse(srcUri)

		artefacts = append(artefacts, &specs.AutogenArtefact{
			SrcUri: []string{srcUri},
			Use:    "",
			Name:   path.Base(uri.Path),
		})

		values["artefacts"] = artefacts
	}

	return g.BaseGenerator.setVersion(atom, version, mapref)
}

func (g *JsonGenerator) Process(atom *specs.AutogenAtom) (*map[string]interface{}, error) {
	log := logger.GetDefaultLogger()
	ans := make(map[string]interface{}, 0)
	var rexclude *regexp.Regexp = nil

	if atom.Json.Url == "" {
		return nil, fmt.Errorf("[%s] No json.url defined!", atom.Name)
	}

	if atom.Json.FilterVersion == "" {
		return nil, fmt.Errorf("[%s] No json.version_filter defined!", atom.Name)
	}

	if atom.Json.Exclude != "" {
		rexclude = regexp.MustCompile(atom.Json.Exclude)
		if rexclude == nil {
			return nil, fmt.Errorf("[%s] invalid regex on exclude", atom.Name)
		}
	}

	method := "GET"
	if atom.Json.Method != "" {
		method = atom.Json.Method
	}

	vars := atom.Vars
	vars["pn"] = atom.Name

	jsonUrl, err := helpers.RenderContentWithTemplates(
		atom.Json.Url,
		"", "", "json.url", vars, []string{},
	)

	uri, err := url.Parse(jsonUrl)
	if err != nil {
		return nil, err
	}

	ssl := false

	if uri.Scheme == "ftp" {
		return nil, fmt.Errorf("Not yet implemented")
	}

	if uri.Scheme == "https" {
		ssl = true
	}

	node := guard_specs.NewRestNode(uri.Host,
		uri.Host+path.Dir(uri.Path), ssl)
	resource := path.Base(uri.Path)

	service := g.GetRestGuardService(uri.Host)
	service.AddNode(node)

	t := service.GetTicket()
	defer t.Rip()

	req, err := g.RestGuard.CreateRequest(t, method, "/"+resource)
	if err != nil {
		return nil, err
	}

	if len(atom.Json.Params) > 0 {
		q := req.URL.Query()

		for k, v := range atom.Json.Params {
			renderedValue, err := helpers.RenderContentWithTemplates(
				v,
				"", "", "param."+k, vars, []string{},
			)
			if err != nil {
				return nil, fmt.Errorf("[%s] error on render query params %s: %s",
					atom.Name, k, err.Error())
			}
			q.Add(k, renderedValue)
		}

		req.URL.RawQuery = q.Encode()

		log.DebugC(fmt.Sprintf(
			":brain:[%s] Using url %s with params %s...", atom.Name, jsonUrl,
			req.URL.RawQuery))
	} else {
		log.DebugC(fmt.Sprintf(
			":brain:[%s] Using url %s...", atom.Name, jsonUrl))
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

	bytesResp, err := ioutil.ReadAll(t.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("error on read body: %s", err.Error())
	}

	obj, err := oj.ParseString(string(bytesResp))
	if err != nil {
		return nil, fmt.Errorf("error on parsing json string: %s", err.Error())
	}

	ans["json_body"] = obj

	// Parse version filter
	versionFilter, err := jp.ParseString(atom.Json.FilterVersion)
	if err != nil {
		return nil, fmt.Errorf("error on parsing version_filter: %s", err.Error())
	}

	jsonVersions := versionFilter.Get(obj)

	versions := []string{}

	for _, v := range jsonVersions {
		vstr, valid := v.(string)
		if !valid {
			vmap, valid := v.(map[string]any)
			if !valid {
				return nil, fmt.Errorf("error on convert filter version element in string")
			}
			for k := range vmap {
				if rexclude != nil && rexclude.MatchString(k) {
					log.DebugC(fmt.Sprintf(
						":brain:[%s] Version %s excluded.", atom.Name, k))
					continue
				}
				versions = append(versions, k)
			}
			continue
		}
		if rexclude != nil && rexclude.MatchString(vstr) {
			log.DebugC(fmt.Sprintf(
				":brain:[%s] Version %s excluded.", atom.Name, vstr))
			continue
		}
		versions = append(versions, vstr)
	}

	ans["versions"] = versions

	return &ans, nil
}
