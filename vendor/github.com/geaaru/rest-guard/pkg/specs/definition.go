/*
Copyright © 2021 Funtoo Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"io"
	"net/http"
)

type RestTicket struct {
	Id          string         `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`
	Request     *http.Request  `json:"req,omitempty" yaml:"req,omitempty" mapstructure:"req,omitempty"`
	Response    *http.Response `json:"resp,omitempty" yaml:"resp,omitempty" mapstructure:"resp,omitempty"`
	Path        string         `json:"path,omitempty" yaml:"path,omitempty" mapstructure:"path,omitempty"`
	Retries     int            `json:"retries,omitempty" yaml:"retries,omitempty" mapstructure:"retries,omitempty"`
	Service     *RestService   `json:"service,omitempty" yaml:"service,omitempty" mapstructure:"service,omitempty"`
	Node        *RestNode      `json:"node,omitempty" yaml:"node,omitempty" mapstructure:"node,omitempty"`
	FailedNodes RestNodes      `json:"failed_nodes,omitempty" yaml:"failed_nodes,omitempty" mapstructure:"failed_nodes,omitempty"`

	RequestBodyCb  func(t *RestTicket) (bool, io.ReadCloser, error) `json:"-" yaml:"-" mapstructure:"-"`
	RequestCloseCb func(t *RestTicket)                              `json:"-" yaml:"-" mapstructure:"-"`
	Closure        map[string]interface{}                           `json:"-" yaml:"-" mapstructure:"-"`
}

type RestNode struct {
	Name    string `json:"name" yaml:"name" mapstructure:"name"`
	Disable bool   `json:"disable,omitempty" yaml:"disable,omitempty" mapstructure:"disable,omitempty"`
	BaseUrl string `json:"base_url" yaml:"base_url" mapstructure:"base_url"`
	Schema  string `json:"schema,omitempty" yaml:"schema,omitempty" mapstructure:"schema,omitempty"`
	Ssl     bool   `json:"ssl,omitempty" yaml:"ssl,omitempty" mapstructure:"ssl,omitempty"`
}

type RestNodes []*RestNode

type RestService struct {
	Name            string      `json:"name" yaml:"name" mapstructure:"name"`
	Nodes           []*RestNode `json:"nodes" yaml:"nodes" mapstructure:"nodes"`
	Retries         int         `json:"retries,omitempty" yaml:"retries,omitempty" mapstructure:"retries,omitempty"`
	RetryIntervalMs int         `json:"retry_interval_ms,omitempty" yaml:"retry_interval_ms,omitempty" mapstructure:"retry_interval_ms,omitempty"`

	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options,omitempty"`

	RespValidatorCb func(t *RestTicket) (bool, error) `json:"-" yaml:"-" mapstructure:"-"`
}

type RestGuardConfig struct {
	UserAgent string `json:"user_agent,omitempty" yaml:"user_agent,omitempty" mapstructure:"user_agent,omitempty,omitempty"`

	ReqsTimeout         int  `json:"reqs_timeout,omitempty" yaml:"reqs_timeout,omitempty" mapstructure:"reqs_timeout,omitempty"`
	MaxIdleConns        int  `json:"max_idle_conns,omitempty" yaml:"max_idle_conns,omitempty" mapstructure:"max_idle_conns,omitempty"`
	IdleConnTimeout     int  `json:"idle_conn_timeout,omitempty" yaml:"idle_conn_timeout,omitempty" mapstructure:"idle_conn_timeout,omitempty"`
	MaxConnsPerHost     int  `json:"max_conns4host,omitempty" yaml:"max_conns4host,omitempty" mapstructure:"max_conns4host,omitempty"`
	MaxIdleConnsPerHost int  `json:"max_idleconns4host,omitempty" yaml:"max_idleconns4host,omitempty" mapstructure:"max_idleconns4host,omitempty"`
	DisableCompression  bool `json:"disable_compression,omitempty" yaml:"disable_compression,omitempty" mapstructure:"disable_compression,omitempty"`
	InsecureSkipVerify  bool `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty" mapstructure:"insecure_skip_verify,omitempty"`
}

type RestArtefact struct {
	Path    string `json:"path,omitempty" yaml:"path,omitempty" mapstructure:"path,omitempty"`
	Size    int64  `json:"size,omitempty" yaml:"size,omitempty" mapstructure:"size,omitempty"`
	Md5     string `json:"md5,omitempty" yaml:"md5,omitempty" mapstructure:"md5,omitempty"`
	Sha512  string `json:"sha512,omitempty" yaml:"sha512,omitempty" mapstructure:"sha512,omitempty"`
	Blake2b string `json:"blake2b,omitempty" yaml:"blake2b,omitempty" mapstructure:"blake2b,omitempty"`
}
