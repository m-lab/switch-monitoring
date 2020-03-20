package netconf

import (
	"testing"

	"github.com/scottdware/go-junos"
)

type mockConn struct{}

func (mockConn) GetConfig(string, ...string) (string, error) {
	return "Not implemented.", nil
}

func TestNew(t *testing.T) {
	auth := &junos.AuthMethod{}
	netconf := New(auth)
	if netconf.auth != auth {
		t.Errorf("New() didn't return the expected struct.")
	}
}

func TestClient_GetConfigHash(t *testing.T) {
	netconf := New(&junos.AuthMethod{})

	oldConnectFunc := connect
	connect = func(h string, auth *junos.AuthMethod) (connection, error) {
		return &mockConn{}, nil
	}
	netconf.GetConfigHash("test")
	connect = oldConnectFunc
}
