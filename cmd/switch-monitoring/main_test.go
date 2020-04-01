package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/m-lab/go/osx"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/scottdware/go-junos"
	"github.com/stretchr/testify/assert"
)

//
// Mocks used in the subsequent unit tests.
//

type mockNetconf struct {
	// How many times GetConfigHash has been called.
	getConfigCalled int
	mustFail        bool
}

func (n *mockNetconf) GetConfig(hostname string, section ...string) (string, error) {
	n.getConfigCalled++
	if n.mustFail {
		return "", fmt.Errorf("error")
	}
	return "not implemented", nil
}

type mockHTTPProvider struct {
	// How many times Get has been called.
	getCalled    int
	mustFail     bool
	responseBody string
}

func (prov *mockHTTPProvider) Get(string) (*http.Response, error) {
	prov.getCalled++
	if prov.mustFail {
		return nil, fmt.Errorf("error")
	}
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(prov.responseBody)),
		StatusCode: http.StatusOK,
	}, nil
}

//
// Tests.
//

func Test_main(t *testing.T) {
	assert := assert.New(t)
	netconf := &mockNetconf{}
	siteinfo := &mockHTTPProvider{
		responseBody: `{"abc01": {}}`,
	}

	oldNewNetconf := newNetconf
	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf
	}

	oldHTTPClient := httpClient
	httpClient = func(timeout time.Duration) internal.HTTPProvider {
		return siteinfo
	}

	// Replace osExit so that tests don't stop running.
	osExit = func(code int) {
		if code != 1 {
			t.Fatalf("Expected a 1 exit code, got %d.", code)
		}

		panic("os.Exit called")
	}

	defer func() {
		osExit = os.Exit
	}()

	// If no SSH key is provided, main() shoud fail.
	assert.PanicsWithValue("os.Exit called", main,
		"os.Exit was not called")

	restore := osx.MustSetenv("KEY", "/path/to/key")

	main()
	if netconf.getConfigCalled == 0 {
		t.Errorf("GetConfig() has not been called.")
	}

	if siteinfo.getCalled == 0 {
		t.Errorf("Get() has not been called.")
	}

	// Make GetConfig() fail.
	netconf.mustFail = true
	main()
	netconf.mustFail = false

	restore()
	newNetconf = oldNewNetconf
	httpClient = oldHTTPClient

}

func Test_newNetconf(t *testing.T) {
	netconf := newNetconf(&junos.AuthMethod{})
	if netconf == nil {
		t.Errorf("newNetconf() returned nil.")
	}
}

func Test_httpClient(t *testing.T) {
	client := httpClient(0)
	if client == nil {
		t.Errorf("httpClient() returned nil.")
	}
}

func Test_switches(t *testing.T) {
	siteinfo := &mockHTTPProvider{}

	oldHTTPClient := httpClient
	httpClient = func(timeout time.Duration) internal.HTTPProvider {
		return siteinfo
	}

	siteinfo.responseBody = `{"abc01": {}}`
	res, err := switches("test")
	if err != nil {
		t.Errorf("switches() returned err: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("switches(): expected one string, found %v", len(res))
	}

	// Get() fails.
	siteinfo.mustFail = true
	_, err = switches("test")
	if err == nil {
		t.Errorf("switches(): expected err, got nil.")
	}
	siteinfo.mustFail = false

	// No content.
	siteinfo.responseBody = ``
	_, err = switches("test")
	if err == nil {
		t.Errorf("switches(): expected err, got nil.")
	}

	// JSON is an empty object.
	siteinfo.responseBody = `{}`
	_, err = switches("test")
	if err == nil {
		t.Errorf("switches(): expected err, got nil.")
	}

	httpClient = oldHTTPClient
}
