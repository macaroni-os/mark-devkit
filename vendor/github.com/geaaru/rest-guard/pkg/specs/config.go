/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
)

const (
	RGuardVersion = "0.4.0"
)

func NewConfig() *RestGuardConfig {
	return &RestGuardConfig{
		UserAgent: fmt.Sprintf("RestGuard v%s", RGuardVersion),

		ReqsTimeout:         120,
		MaxIdleConns:        10,
		IdleConnTimeout:     30,
		MaxConnsPerHost:     5,
		MaxIdleConnsPerHost: 30,
		DisableCompression:  false,
		InsecureSkipVerify:  false,
	}
}
