package collector

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/m-lab/go/rtx"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Mocks for content and netconf providers.
type contentProvider struct {
	filepath string
	fail     bool
}

func (c *contentProvider) Get(context.Context) ([]byte, error) {
	if c.fail {
		return nil, fmt.Errorf("GetConfig error")
	}
	content, err := ioutil.ReadFile(c.filepath)
	rtx.Must(err, "Cannot read test data")
	return content, nil
}

type netconfProvider struct {
	filepath string
	fail     bool
}

func (n *netconfProvider) GetConfig(hostname string, sections ...string) (string, error) {
	if n.fail {
		return "", fmt.Errorf("GetConfig error")
	}
	content, err := ioutil.ReadFile(n.filepath)
	rtx.Must(err, "Cannot read test data")
	return string(content), nil
}

func TestNew(t *testing.T) {
	if New("s1.abc01.measurement-lab.org", Config{
		ProjectID: "test",
		Netconf:   nil,
		Provider:  nil,
	}) == nil {
		t.Errorf("New() returned nil.")
	}
}

func TestConfigCheckerCollector_Collect(t *testing.T) {
	metadata := `# HELP switch_monitoring_config_match Configuration check result for this target
# TYPE switch_monitoring_config_match gauge
`
	netconf := &netconfProvider{
		filepath: "testdata/abc01.conf",
	}

	provider := &contentProvider{
		filepath: "testdata/abc01.conf",
	}

	config := Config{
		ProjectID: "test",
		Netconf:   netconf,
		Provider:  provider,
	}

	collector := New("s1.abc01.measurement-lab.org", config)

	expected := metadata + `
switch_monitoring_config_match{status="ok",target="s1.abc01.measurement-lab.org"} 1
`
	err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
	if err != nil {
		t.Errorf("Collect() returned err: %v", err)
	}

	// Compare two different configs.
	expected = metadata + `
switch_monitoring_config_match{status="config_mismatch",target="s1.abc01.measurement-lab.org"} 1
`
	provider.filepath = "testdata/abc02.conf"
	err = testutil.CollectAndCompare(collector, strings.NewReader(expected))
	if err != nil {
		t.Errorf("Collect() returned err: %v", err)
	}

	// Make the content provider fail.
	expected = metadata + `
switch_monitoring_config_match{status="config_not_found_gcs",target="s1.abc01.measurement-lab.org"} 1
`
	provider.fail = true
	err = testutil.CollectAndCompare(collector, strings.NewReader(expected))
	if err != nil {
		t.Errorf("Collect() returned err: %v", err)
	}
	provider.fail = false

	// Make netconf fail.
	expected = metadata + `
switch_monitoring_config_match{status="config_not_found_switch",target="s1.abc01.measurement-lab.org"} 1
`
	netconf.fail = true
	err = testutil.CollectAndCompare(collector, strings.NewReader(expected))
	if err != nil {
		t.Errorf("Collect() returned err: %v", err)
	}
	netconf.fail = false
}
