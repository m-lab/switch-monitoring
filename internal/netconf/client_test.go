package netconf

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/scottdware/go-junos"
)

type mockConnector struct {
	mustFail     bool
	mustFailConn bool
}

func (c mockConnector) NewSession(string, *junos.AuthMethod) (connection, error) {
	if c.mustFail {
		return nil, fmt.Errorf("error")
	}
	return &mockConnection{
		mustFail: c.mustFailConn,
	}, nil
}

type mockConnection struct {
	mustFail bool
}

func (c mockConnection) GetConfig(string, ...string) (string, error) {
	if c.mustFail {
		return "", fmt.Errorf("error")
	}

	// Read test file.
	testfile, err := ioutil.ReadFile("testdata/abc01_qfx5100.conf")
	if err != nil {
		return "", err
	}
	return string(testfile), nil
}

func TestNew(t *testing.T) {
	auth := &junos.AuthMethod{}
	netconf := New(auth)
	if netconf.auth != auth {
		t.Errorf("New() didn't return the expected struct.")
	}
}

func TestClient_GetConfigHash(t *testing.T) {
	mockConnector := &mockConnector{}
	netconf := &Client{
		auth:      &junos.AuthMethod{},
		connector: mockConnector,
	}

	res, err := netconf.GetConfig("test")
	if err != nil {
		t.Errorf("GetConfig(): expected nil, got %v", err)
	}

	// Check the content has been cleaned as expected.
	if strings.HasPrefix(res, "#") {
		t.Errorf("GetConfig(): comments have not been removed.")
	}

	if !strings.HasPrefix(res, "version") {
		t.Errorf("GetConfig(): config does not begin with 'version'")
	}

	// Let the connector fail.
	mockConnector.mustFail = true
	_, err = netconf.GetConfig("test")
	if err == nil {
		t.Errorf("GetConfig(): expected err, got nil.")
	}
	mockConnector.mustFail = false

	// Let connection.GetConfig() fail.
	mockConnector.mustFailConn = true
	_, err = netconf.GetConfig("test")
	if err == nil {
		t.Errorf("GetConfig(): expected err, got nil.")
	}

}
