package netconf

import "github.com/scottdware/go-junos"

// These types provide an abstraction for the underlying connector and
// connection to a NETCONF-enabled device, so that they can be unit tested.

type connection interface {
	GetConfig(string, ...string) (string, error)
}

type connector interface {
	NewSession(string, *junos.AuthMethod) (connection, error)
}

type junosConnector struct{}

func (junosConnector) NewSession(host string, auth *junos.AuthMethod) (connection, error) {
	return junos.NewSession(host, auth)
}
