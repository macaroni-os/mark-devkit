/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"strings"

	"github.com/macaroni-os/mark-devkit/pkg/logger"
	"github.com/macaroni-os/mark-devkit/pkg/specs"

	"github.com/geaaru/rest-guard/pkg/guard"
	guard_specs "github.com/geaaru/rest-guard/pkg/specs"
)

type NotifierDiscord struct {
	RestGuard   *guard.RestGuard
	MapServices map[string]*guard_specs.RestService
	RateLimit   string
	Hook        *specs.MarkDevkitHook
}

type WebhookMsg struct {
	Content string `json:"content"`
}

func NewNotifierDiscord(opts map[string]string, hook *specs.MarkDevkitHook) *NotifierDiscord {
	log := logger.GetDefaultLogger()
	rg, _ := guard.NewRestGuard(log.Config.GetRest())

	ans := &NotifierDiscord{
		RestGuard: rg,
		Hook:      hook,
	}

	// Set storage
	storage := *log.Config.GetStorage()
	mServicesI, ok := storage[specs.NotifyDiscord]
	if ok {
		ans.MapServices, _ = mServicesI.(map[string]*guard_specs.RestService)
	} else {
		// POST: storage is not initialized.
		ans.MapServices = make(map[string]*guard_specs.RestService, 0)
		storage[specs.NotifyDiscord] = ans.MapServices
	}

	if limit, present := opts[guard_specs.ServiceRateLimiter]; present {
		log.DebugC(fmt.Sprintf(
			":brain: Using rate limit %s...", limit))
		ans.RateLimit = limit
	}

	return ans
}

func (n *NotifierDiscord) GetRestGuardService(service string) *guard_specs.RestService {
	var ans *guard_specs.RestService = nil
	if s, present := n.MapServices[service]; present {
		ans = s.Clone()
		// Ensure same rate limiter
		ans.RateLimiter = s.RateLimiter
	} else {
		ans = guard_specs.NewRestService(service)
		ans.Retries = 3
		ans.SetOption(guard_specs.ServiceRateLimiter, n.RateLimit)
		ans.SetRateLimiter()

		// Temporary until a new default Restguard callback is available.
		// Discord return 204 on ok.
		respValidator := func(t *guard_specs.RestTicket) (bool, error) {
			ans := false
			if t.Response != nil &&
				(t.Response.StatusCode == 200 || t.Response.StatusCode == 201 || t.Response.StatusCode == 204) {
				ans = true
			}
			return ans, nil
		}
		ans.RespValidatorCb = respValidator
	}

	return ans
}

func (n *NotifierDiscord) GetType() string {
	return specs.NotifyDiscord
}

func (n *NotifierDiscord) GetHook() *specs.MarkDevkitHook { return n.Hook }

func (n *NotifierDiscord) Notify(atom *specs.AutogenAtom, msg string) error {
	log := logger.GetDefaultLogger()

	log.DebugC(fmt.Sprintf(
		":brain:[%s] Notify to discord message: %s", atom.Name, msg))

	uri, err := url.Parse(n.Hook.Url)
	if err != nil {
		return fmt.Errorf("error on parse discord url: %s", err.Error())
	}

	ssl := false
	if uri.Scheme == "https" {
		ssl = true
	}

	node := guard_specs.NewRestNode(uri.Host,
		uri.Host+path.Dir(uri.Path), ssl)

	resource := ""
	if !strings.HasSuffix(n.Hook.Url, "/") {
		resource = path.Base(uri.Path)
	}

	webhookContent := WebhookMsg{
		Content: fmt.Sprintf(
			"🤖 [%s] %s",
			atom.Name, msg,
		),
	}

	body, err := json.Marshal(webhookContent)
	if err != nil {
		return err
	}

	service := n.GetRestGuardService(uri.Host)
	service.AddNode(node)

	t := service.GetTicket()
	defer t.Rip()

	fBody := func(t *guard_specs.RestTicket) (bool, io.ReadCloser, error) {
		return true, ioutil.NopCloser(bytes.NewReader(body)), nil
	}
	t.RequestBodyCb = fBody
	_, err = n.RestGuard.CreateRequest(t, "POST", "/"+resource)
	if err != nil {
		return err
	}
	t.Request.Header.Add("Content-Type", "application/json")

	err = n.RestGuard.Do(t)
	if err != nil {
		if t.Response != nil {
			return fmt.Errorf("error on notify %s: %s - %s", atom.Name, err.Error(), t.Response.Status)
		} else {
			return fmt.Errorf("error on notify %s: %s", atom.Name, err.Error())
		}
	}

	return nil
}
