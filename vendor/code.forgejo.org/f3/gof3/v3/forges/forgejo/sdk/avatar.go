// Copyright F3 Authors
// SPDX-License-Identifier: MIT

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type createAvatarOption struct {
	Image string `json:"image"`
}

func (c *Client) CreateAvatar(avatarable, image string) (*Response, error) {
	body, err := json.Marshal(&createAvatarOption{Image: image})
	if err != nil {
		return nil, err
	}
	status, resp, err := c.getStatusCode("POST", fmt.Sprintf("%s/avatar", avatarable), jsonHeader, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}
	if status != 204 {
		return resp, fmt.Errorf("unexpected Status: %d", status)
	}
	return resp, nil
}

func (c *Client) DeleteAvatar(avatarable string) (*Response, error) {
	if err := escapeValidatePathSegments(&avatarable); err != nil {
		return nil, err
	}

	status, resp, err := c.getStatusCode("DELETE", fmt.Sprintf("%s/avatar", avatarable), nil, nil)
	if err != nil {
		return resp, err
	}
	if status != 204 {
		return resp, fmt.Errorf("unexpected Status: %d", status)
	}
	return resp, nil
}
