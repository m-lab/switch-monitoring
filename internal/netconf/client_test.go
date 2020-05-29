package netconf

import (
	"fmt"
	"io/ioutil"
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

func (c *mockConnection) GetConfig(string, ...string) (string, error) {
	if c.mustFail {
		return "", fmt.Errorf("error")
	}

	// Read test file.
	testfile, err := ioutil.ReadFile("testdata/abc01.conf")
	if err != nil {
		return "", err
	}
	return string(testfile), nil
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
	if len(res) == 0 {
		t.Errorf("GetConfig(): len(config) is zero.")
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
