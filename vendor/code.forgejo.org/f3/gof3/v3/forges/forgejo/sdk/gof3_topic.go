// Copyright F3 Authors
// SPDX-License-Identifier: MIT

package sdk

import (
	"fmt"
	"time"
)

type Topic struct {
	ID        int64      `json:"id"`
	Name      string     `json:"topic_name"`
	RepoCount int        `json:"repo_count"`
	Created   *time.Time `json:"created_at"`
	Updated   *time.Time `json:"updated_at"`
}

type SearchTopicsOptions struct {
	ListOptions
	Keyword string
}

func (opt *SearchTopicsOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Keyword != "" {
		query.Add("q", opt.Keyword)
	}
	return query.Encode()
}

func (c *Client) SearchTopics(opt SearchTopicsOptions) ([]*Topic, *Response, error) {
	opt.setDefaults()
	topics := make([]*Topic, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/topics/search?%s", opt.getURLQuery().Encode()), nil, nil, &topics)
	return topics, resp, err
}
