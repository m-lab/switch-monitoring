package main

import (
	"fmt"
	"io/ioutil"
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

//
// Tests.
//

func Test_main(t *testing.T) {
	assert := assert.New(t)
	netconf := &mockNetconf{
		configFile: "testdata/abc01.conf",
	}

	oldNewNetconf := newNetconf
	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf
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

	restoreKey := osx.MustSetenv("SSH_KEY", "/path/to/key")
	restorePort := osx.MustSetenv("LISTENADDR", ":0")

	go main()

	time.Sleep(1 * time.Second)
	cancel()

	restoreKey()
	restorePort()
	newNetconf = oldNewNetconf
}

func Test_newNetconf(t *testing.T) {
	netconf := newNetconf(&junos.AuthMethod{})
	if netconf == nil {
		t.Errorf("newNetconf() returned nil.")
	}
}
