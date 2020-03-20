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

const responseBody = `{"abc01": {}}`

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
	getCalled int
}

func (prov *mockHTTPProvider) Get(string) (*http.Response, error) {
	prov.getCalled++
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(responseBody)),
		StatusCode: http.StatusOK,
	}, nil
}

// Tests.
func Test_main(t *testing.T) {
	assert := assert.New(t)
	netconf := &mockNetconf{}
	siteinfo := &mockHTTPProvider{}

	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf
	}

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
	restore()

	if netconf.getConfigCalled == 0 {
		t.Errorf("GetConfig() has not been called.")
	}

	if siteinfo.getCalled == 0 {
		t.Errorf("Get() has not been called.")
	}
}
