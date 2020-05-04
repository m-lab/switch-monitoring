package netconf

import (
	"testing"

	"github.com/scottdware/go-junos"
	"golang.org/x/crypto/ssh"
)

func Test_junosConnector_NewSession(t *testing.T) {
	// Let NewSession fail due to an empty AuthMethod.
	j := &junosConnector{}
	_, err := j.NewSession("", &junos.AuthMethod{})
	if err == nil {
		t.Errorf("NewSession(): expected err, got nil.")
	}

	// Attempt to use a non-existing key file.
	auth := &junos.AuthMethod{
		PrivateKey: "thiswillfail",
		Username:   "thiswillfail",
	}

	_, err = j.NewSession("", auth)
	if err == nil {
		t.Errorf("NewSession() expected err, got nil.")
	}

	// This should succeed.
	auth.PrivateKey = "testdata/dummy.key"
	oldNewSession := newSession
	newSession = func(host string, clientConfig *ssh.ClientConfig) (*junos.Junos, error) {
		return &junos.Junos{}, nil
	}
	s, err := j.NewSession("", auth)
	if err != nil {
		t.Errorf("NewSession() expected err, got nil.")
	}
	if s == nil {
		t.Errorf("NewSession() returned nil.")
	}
	newSession = oldNewSession
}
