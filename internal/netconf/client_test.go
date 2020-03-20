package netconf

import (
	"fmt"
	"testing"

	"github.com/scottdware/go-junos"
)

type mockConnector struct {
	mustFail bool
}

func (c mockConnector) NewSession(string, *junos.AuthMethod) (connection, error) {
	if c.mustFail {
		return nil, fmt.Errorf("error")
	}
	return &mockConnection{}, nil
}

type mockConnection struct{}

func (mockConnection) GetConfig(string, ...string) (string, error) {
	return "test", nil
}

func TestNew(t *testing.T) {
	auth := &junos.AuthMethod{}
	netconf := New(auth)
	if netconf.auth != auth {
		t.Errorf("New() didn't return the expected struct.")
	}
}

func TestClient_GetConfigHash(t *testing.T) {
	netconf := &Client{
		auth:      &junos.AuthMethod{},
		connector: &mockConnector{},
	}

	res, err := netconf.GetConfig("test")
	if err != nil {
		t.Errorf("GetConfig(): expected nil, got %v", err)
	}

	if res != "test" {
		t.Errorf("GetConfig(): unexpected value %v", res)
	}
}
