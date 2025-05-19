/*
Copyright © 2021 Funtoo Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func defaultRespCheck(t *RestTicket) (bool, error) {
	ans := false
	if t.Response != nil &&
		(t.Response.StatusCode == 200 || t.Response.StatusCode == 201) {
		ans = true
	}
	return ans, nil
}

func NewRestService(n string) *RestService {
	return &RestService{
		Name:            n,
		Nodes:           []*RestNode{},
		Retries:         0,
		RespValidatorCb: defaultRespCheck,
		RetryIntervalMs: 10,
		RateLimiter:     nil,
		Options:         make(map[string]string, 0),
	}
}

func (s *RestService) GetName() string  { return s.Name }
func (s *RestService) SetName(n string) { s.Name = n }
func (s *RestService) AddNode(n *RestNode) {
	s.Nodes = append(s.Nodes, n)
}

func (s *RestService) GetNodes() []*RestNode {
	return s.Nodes
}

func (s *RestService) HasOption(k string) bool {
	_, isPresent := s.Options[k]
	return isPresent
}

func (s *RestService) GetOption(k string) (string, error) {
	v, ok := s.Options[k]
	if !ok {
		return "", errors.New("Key not found")
	}
	return v, nil
}

func (s *RestService) HasRateLimiter() bool {
	return s.RateLimiter != nil
}

func (s *RestService) GetRateLimiter() *rate.Limiter { return s.RateLimiter }

func (s *RestService) SetRateLimiter() error {
	v, err := s.GetOption(ServiceRateLimiter)
	if err != nil {
		return fmt.Errorf("No rate limits option available")
	}

	// Setup a new Rate Limiter
	reqs, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("Invalid rate limits option: %s", err.Error())
	}

	// N reqs every second
	s.RateLimiter = rate.NewLimiter(rate.Every(1*time.Second), reqs)
	return nil
}

func (s *RestService) SetOption(k, v string) {
	s.Options[k] = v
}

func (s *RestService) GetTicket() *RestTicket {
	ans := &RestTicket{
		Id:          uuid.New().String(),
		Retries:     0,
		Node:        nil,
		Path:        "",
		Service:     s,
		Closure:     make(map[string]interface{}, 0),
		FailedNodes: []*RestNode{},
	}

	return ans
}

func (s *RestService) Clone() *RestService {
	ans := &RestService{
		Name:            s.Name,
		Retries:         s.Retries,
		RespValidatorCb: s.RespValidatorCb,
		RetryIntervalMs: s.RetryIntervalMs,
		Options:         make(map[string]string, 0),
	}

	for k, v := range s.Options {
		ans.Options[k] = v

		if k == ServiceRateLimiter {
			// Ignoring error. I assume that the value is correct.
			ans.SetRateLimiter()
		}
	}

	return ans
}
