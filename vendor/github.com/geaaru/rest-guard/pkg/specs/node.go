/*
Copyright © 2021 Funtoo Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"strings"
)

func NewRestNode(name, burl string, ssl bool) *RestNode {
	if strings.HasSuffix(burl, "/") {
		burl = burl[0 : len(burl)-1]
	}
	return &RestNode{
		Name:    name,
		BaseUrl: burl,
		Ssl:     ssl,
		Disable: false,
	}
}

func (n *RestNode) IsActive() bool    { return !n.Disable }
func (n *RestNode) SetDisable(b bool) { n.Disable = b }

func (n *RestNode) GetUrlPrefix() string {
	ans := ""
	if n.Ssl {
		ans = "https://"
	} else {
		ans = "http://"
	}

	ans += n.BaseUrl

	return ans
}

func (n *RestNode) Equal(o *RestNode) bool {
	return n.Name == o.Name && n.Ssl == o.Ssl && n.BaseUrl == o.BaseUrl
}

func (nn RestNodes) HasNode(n *RestNode) bool {
	for _, e := range nn {
		if e.Equal(n) {
			return true
		}
	}
	return false
}
