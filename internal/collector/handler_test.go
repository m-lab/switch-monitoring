package collector

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/m-lab/go/content"
)

func TestNewHandler(t *testing.T) {
	netconf := &netconfProvider{}
	h := NewHandler("test", netconf)
	if h == nil || h.netconf != netconf || h.projectID != "test" {
		t.Errorf("NewHandler() didn't return the expected value")
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	metadata := `# HELP switch_monitoring_config_match Configuration check result for this target
# TYPE switch_monitoring_config_match gauge
`

	tests := []struct {
		name              string
		r                 *http.Request
		status            int
		body              string
		getConfigMustFail bool
	}{
		{
			name: "ok-configs-match",
			r: httptest.NewRequest("GET",
				"/v1/check?target=s1.abc01.measurement-lab.org", nil),
			status: http.StatusOK,
			body: metadata + `switch_monitoring_config_match{status="ok",` +
				`target="s1.abc01.measurement-lab.org"} 1
`,
		},
		{
			name:   "method-not-allowed",
			r:      httptest.NewRequest("POST", "/v1/check", nil),
			status: http.StatusMethodNotAllowed,
		},
		{
			name:   "target-not-provided",
			r:      httptest.NewRequest("GET", "/v1/check", nil),
			status: http.StatusBadRequest,
			body:   "URL parameter 'target' is missing",
		},
		{
			name:   "invalid-target",
			r:      httptest.NewRequest("GET", "/v1/check?target=invalid", nil),
			status: http.StatusBadRequest,
			body:   "cannot extract site from hostname: invalid",
		},
		{
			name:              "failure-getting-content-provider",
			r:                 httptest.NewRequest("GET", "/v1/check?target=s1.abc01", nil),
			status:            http.StatusInternalServerError,
			getConfigMustFail: true,
		},
	}

	netconf := &netconfProvider{
		filepath: "testdata/abc01.conf",
	}

	provider := &contentProvider{
		filepath: "testdata/abc01.conf",
	}

	handler := NewHandler("test", netconf)
	for _, test := range tests {
		rr := httptest.NewRecorder()

		handler.getConfigFunc = func(context.Context, *url.URL) (content.Provider, error) {
			if test.getConfigMustFail {
				return nil, fmt.Errorf("getConfigFunc() error")
			}
			return provider, nil
		}
		handler.ServeHTTP(rr, test.r)

		resp := rr.Result()

		if resp.StatusCode != test.status {
			t.Errorf("ServeHTTP - expected %d, got %d", test.status,
				resp.StatusCode)
		}

		if test.body != "" {
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Errorf("ServeHTTP() - cannot read response: %v", err)
			}
			if string(body) != test.body {
				t.Errorf("ServeHTTP() - unexpected response: \n%s", string(body))
			}
		}
	}
}

func TestHandler_getProviderForConfig(t *testing.T) {
	h := NewHandler("test", &netconfProvider{})
	oldParseURL := parseURL
	parseURL = func(rawurl string) (*url.URL, error) {
		return nil, fmt.Errorf("parseURL() error")
	}
	_, err := h.getProviderForConfig("://")
	if err == nil {
		t.Errorf("getProviderForConfig(): expected err, got nil.")
	}

	parseURL = oldParseURL

}
