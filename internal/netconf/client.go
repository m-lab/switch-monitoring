package netconf

import (
	"github.com/scottdware/go-junos"
)

// Client is a client to get the switch configuration using the
// NETCONF protocol.
type Client struct {
	auth      *junos.AuthMethod
	connector connector
}

// New returns a new NetconfClient.
func New(auth *junos.AuthMethod) Client {
	return Client{
		auth:      auth,
		connector: junosConnector{},
	}
}

// GetConfig connects to a switch, reads the specified configuration section
// and returns its content. The section can be an empty string. In that case,
// the whole configuration will be read.
func (c Client) GetConfig(hostname string, section ...string) (string, error) {
	jnpr, err := c.connector.NewSession(hostname, c.auth)
	if err != nil {
		return "", err
	}

	config, err := jnpr.GetConfig("text", section...)
	if err != nil {
		return "", err
	}

	return config, nil
}
