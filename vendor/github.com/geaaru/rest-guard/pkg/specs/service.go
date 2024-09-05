/*
Copyright © 2021 Funtoo Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"errors"

	"github.com/google/uuid"
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
	}

	return ans
}
