package netconf

import (
	"errors"
	"fmt"
	"testing"

	"github.com/scottdware/go-junos"
)

type mockConnector struct {
	mustFail       bool
	mustFailConfig bool
	mustFailDiff   bool
}

func (c mockConnector) NewSession(string, *junos.AuthMethod) (connection, error) {
	if c.mustFail {
		return nil, fmt.Errorf("error")
	}
	return &mockConnection{
		mustFailConfig: c.mustFailConfig,
		mustFailDiff:   c.mustFailDiff,
	}, nil
}

type mockConnection struct {
	mustFailConfig bool
	mustFailDiff   bool
}

func (c *mockConnection) Config(interface{}, string, bool) error {
	if c.mustFailConfig {
		return errors.New("error")
	}

	return nil
}

func (c *mockConnection) Diff(int) (string, error) {
	if c.mustFailDiff {
		return "", errors.New("error")
	}

	return "", nil
}

func (c *mockConnection) Close() {
	// not implemented.
}

func TestNew(t *testing.T) {
	auth := &junos.AuthMethod{}
	netconf := New(auth)
	if netconf.auth != auth {
		t.Errorf("New() didn't return the expected struct.")
	}
}
