package internal

import (
	"context"
	"net/http"
)

// This file defines interfaces to allow for object mocking in unit tests.

// NetconfClient is a generic NETCONF client.
type NetconfClient interface {
	GetConfig(hostname string, section ...string) (string, error)
}

// HTTPProvider is a data provider returning HTTP responses.
// http.Client satisfies this interface.
type HTTPProvider interface {
	Get(string) (*http.Response, error)
}

// ConfigProvider is a configuration file provider returning a configuration
// file's content.
type ConfigProvider interface {
	Get(ctx context.Context) ([]byte, error)
}
