/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package guard

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/geaaru/rest-guard/pkg/specs"
)

type RestGuard struct {
	Client    *http.Client `json:"-" yaml:"-"`
	UserAgent string       `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`

	Services map[string]*specs.RestService `json:"services" yaml:"services"`
	RetryCb  func(guard *RestGuard, t *specs.RestTicket) (*specs.RestNode, error)
}

func NewRestGuard(cfg *specs.RestGuardConfig) (*RestGuard, error) {
	idleConnTimeout, err := time.ParseDuration(fmt.Sprintf("%ds",
		cfg.IdleConnTimeout))
	if err != nil {
		return nil, err
	}

	reqsTimeout, err := time.ParseDuration(fmt.Sprintf("%ds",
		cfg.ReqsTimeout))
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        cfg.MaxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		DisableCompression:  cfg.DisableCompression,
	}

	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	ans := &RestGuard{
		UserAgent: cfg.UserAgent,
		RetryCb:   nil,
		Services:  make(map[string]*specs.RestService, 0),
	}

	ans.Client = &http.Client{
		Transport: transport,
		Timeout:   reqsTimeout,
	}

	return ans, nil
}

func (g *RestGuard) AddRestNode(srv string, n *specs.RestNode) error {
	_, ok := g.Services[srv]
	if !ok {
		return errors.New("Service " + srv + " not found")
	}

	g.Services[srv].AddNode(n)

	return nil
}

func (g *RestGuard) AddService(srv string, s *specs.RestService) {
	g.Services[srv] = s
}

func (g *RestGuard) GetUserAgent() string { return g.UserAgent }

func (g *RestGuard) GetService(srv string) (*specs.RestService, error) {
	s, ok := g.Services[srv]
	if !ok {
		return nil, errors.New("Service " + srv + " not found")
	}
	return s, nil
}

func (g *RestGuard) CreateRequest(t *specs.RestTicket, method, path string) (*http.Request, error) {

	if t.Service == nil {
		return nil, errors.New("The ticket is without service.")
	}

	if len(t.Service.Nodes) == 0 {
		return nil, errors.New("The service is without nodes.")
	}

	if t.Service.RespValidatorCb == nil {
		return nil, errors.New("Service without response validator")
	}

	activeNodes := []*specs.RestNode{}
	for idx := range t.Service.Nodes {
		if t.Service.Nodes[idx].Disable {
			continue
		}
		activeNodes = append(activeNodes, t.Service.Nodes[idx])
	}

	if len(activeNodes) == 0 {
		return nil, errors.New("The service is without active nodes.")
	}

	if t.Request != nil {
		t.Response = nil
	}
	t.Path = path

	var rn *specs.RestNode
	if t.Node == nil {
		rn = activeNodes[t.Retries%len(activeNodes)]
		t.Node = rn
	} else {
		rn = t.Node
	}

	url := rn.GetUrlPrefix()
	if strings.HasPrefix(path, "/") {
		url += path
	} else {
		url += "/" + path
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if g.GetUserAgent() != "" {
		req.Header.Add("User-Agent", g.GetUserAgent())
	}

	t.Request = req

	if t.RequestBodyCb != nil {
		hasBody, reader, err := t.RequestBodyCb(t)
		if err != nil {
			return nil, err
		}
		if hasBody {
			t.Request.Body = reader
		}
	}

	return req, nil
}

func (g *RestGuard) doClient(c *http.Client, t *specs.RestTicket) error {
	var ans error = nil

	ctx := context.Background()

	handleRetry := func() error {
		t.Retries++
		currReq := t.Request
		t.AddFail(t.Node)
		if g.RetryCb != nil {
			node, err := g.RetryCb(g, t)
			if err != nil {
				return err
			}
			t.Node = node
		} else {
			t.Node = nil
		}
		newReq, err := g.CreateRequest(t, currReq.Method, t.Path)
		if err != nil {
			return err
		}
		newReq.Header = currReq.Header

		if t.FailedNodes.HasNode(t.Node) && t.Service.RetryIntervalMs > 0 {
			sleepms, err := time.ParseDuration(fmt.Sprintf(
				"%dms", t.Service.RetryIntervalMs))
			if err != nil {
				return err
			}
			time.Sleep(sleepms)
		}

		return nil
	}

	var lastResp *http.Response = nil

	for t.Retries <= t.Service.Retries {

		// If rate limiter is present on service
		// ensure limits
		if t.Service.HasRateLimiter() {
			// NOTE: Check if the wait lock requests for all services.
			err := t.Service.GetRateLimiter().Wait(ctx)
			if err != nil {
				return fmt.Errorf("error on rate limiting: %s", err.Error())
			}
		}
		resp, err := c.Do(t.Request)
		t.Response = resp
		lastResp = resp
		if t.RequestCloseCb != nil {
			t.RequestCloseCb(t)
		}
		if err != nil {
			ans = err
			err = handleRetry()
			if err != nil {
				return err
			}
		} else {
			ans = nil
			valid, errValid := t.Service.RespValidatorCb(t)
			if !valid {
				if errValid != nil {
					ans = errValid
				} else {
					ans = errors.New("Received invalid response")
				}
				err = handleRetry()
				if err != nil {
					return err
				}
			} else {
				ans = nil
				break
			}
		}
	}

	if ans != nil {
		t.Response = lastResp
		t.Retries--
	}

	return ans
}

func (g *RestGuard) DoWithTimeout(t *specs.RestTicket, timeoutSec int) error {
	// Could be needed to hava a way to execute HTTP call with a custom
	// timeout. In this case, I create a new client with the timeout
	// in input.

	origTransport := g.Client.Transport.(*http.Transport)

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        origTransport.MaxIdleConns,
		IdleConnTimeout:     origTransport.IdleConnTimeout,
		MaxIdleConnsPerHost: origTransport.MaxIdleConnsPerHost,
		MaxConnsPerHost:     origTransport.MaxConnsPerHost,
	}

	if origTransport.TLSClientConfig != nil &&
		origTransport.TLSClientConfig.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	reqsTimeout, err := time.ParseDuration(fmt.Sprintf("%ds", timeoutSec))
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   reqsTimeout,
	}

	return g.doClient(client, t)
}

func (g *RestGuard) Do(t *specs.RestTicket) error {
	return g.doClient(g.Client, t)
}
