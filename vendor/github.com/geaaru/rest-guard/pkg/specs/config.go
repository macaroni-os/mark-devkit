/*
Copyright © 2024-2026 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
)

const (
	RGuardVersion = "0.8.0"
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
