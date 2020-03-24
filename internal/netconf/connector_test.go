package netconf

import (
	"testing"

	"github.com/scottdware/go-junos"
)

func Test_junosConnector_NewSession(t *testing.T) {
	j := &junosConnector{}
	_, err := j.NewSession("", &junos.AuthMethod{})
	if err == nil {
		t.Errorf("NewSession(): expected err, got nil.")
	}
}
