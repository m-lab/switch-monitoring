package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/m-lab/go/osx"
	"github.com/m-lab/go/rtx"
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
	configFile      string
}

func (n *mockNetconf) GetConfig(hostname string, section ...string) (string, error) {
	n.getConfigCalled++
	if n.mustFail {
		return "", fmt.Errorf("error")
	}
	config, err := ioutil.ReadFile(n.configFile)
	rtx.Must(err, "Cannot read test data")
	return string(config), nil
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

type mockConfigProvider struct {
	configFile string
	mustFail   bool
}

func (c mockConfigProvider) Get(context.Context) ([]byte, error) {
	if c.mustFail {
		return nil, fmt.Errorf("Get() failed")
	}
	config, err := ioutil.ReadFile(c.configFile)
	rtx.Must(err, "Cannot read test data")
	return config, nil
}

//
// Tests.
//

func Test_main(t *testing.T) {
	assert := assert.New(t)
	netconf := &mockNetconf{
		configFile: "testdata/abc01.conf",
	}
	siteinfo := &mockHTTPProvider{
		responseBody: `{"abc01": {}}`,
	}
	configProvider := &mockConfigProvider{
		configFile: "testdata/abc01.conf",
	}

	oldNewNetconf := newNetconf
	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf
	}

	oldHTTPClient := httpClient
	httpClient = func(timeout time.Duration) internal.HTTPProvider {
		return siteinfo
	}

	oldConfigFromURL := configFromURL
	configFromURL = func(ctx context.Context, u *url.URL) (internal.ConfigProvider, error) {
		return configProvider, nil
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
	configFromURL = oldConfigFromURL

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

func Test_configFromURL(t *testing.T) {
	url, err := url.Parse("gs://test/file")
	rtx.Must(err, "Cannot create test URL")
	config, err := configFromURL(context.Background(), url)
	if err != nil {
		t.Errorf("configFromURL() returned err: %v", err)
	}
	if config == nil {
		t.Errorf("httpClient() returned nil.")
	}
}

func Test_sites(t *testing.T) {
	siteinfo := &mockHTTPProvider{}

	oldHTTPClient := httpClient
	httpClient = func(timeout time.Duration) internal.HTTPProvider {
		return siteinfo
	}

	siteinfo.responseBody = `{"abc01": {}}`
	res, err := sites("test")
	if err != nil {
		t.Errorf("sites() returned err: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("sites(): expected one string, found %v", len(res))
	}

	// Get() fails.
	siteinfo.mustFail = true
	_, err = sites("test")
	if err == nil {
		t.Errorf("sites(): expected err, got nil.")
	}
	siteinfo.mustFail = false

	// No content.
	siteinfo.responseBody = ``
	_, err = sites("test")
	if err == nil {
		t.Errorf("sites(): expected err, got nil.")
	}

	// JSON is an empty object.
	siteinfo.responseBody = `{}`
	_, err = sites("test")
	if err == nil {
		t.Errorf("sites(): expected err, got nil.")
	}

	httpClient = oldHTTPClient
}

func Test_checkAll(t *testing.T) {
	// Set up test conditions.
	oldConfigFromURL := configFromURL

	configProvider := &mockConfigProvider{
		configFile: "testdata/abc01.conf",
	}
	configFromURL = func(ctx context.Context, u *url.URL) (internal.ConfigProvider, error) {
		return configProvider, nil
	}

	//
	// Tests begin here.
	//

	tests := []struct {
		name           string
		netconf        internal.NetconfClient
		sites          []string
		configMustFail bool
	}{
		{
			name: "ok",
			netconf: &mockNetconf{
				configFile: "testdata/abc01.conf",
			},
			sites: []string{"abc01"},
		},
		{
			name: "config-mismatch",
			netconf: &mockNetconf{
				configFile: "testdata/abc02.conf",
			},
			sites: []string{"abc02"},
		},
		{
			name: "netconf-get-config-fails",
			netconf: &mockNetconf{
				mustFail: true,
			},
			sites: []string{"abc01"},
		},
		{
			name: "config-provider-fails",
			netconf: &mockNetconf{
				configFile: "testdata/abc01.conf",
			},
			sites:          []string{"abc01"},
			configMustFail: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configProvider.mustFail = tt.configMustFail
			checkAll(tt.netconf, tt.sites)
		})
	}
	configFromURL = oldConfigFromURL
}
