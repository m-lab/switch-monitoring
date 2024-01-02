package netconf

import (
	"fmt"

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

func (c Client) CompareConfig(hostname, config string) (bool, error) {
	jnpr, err := c.connector.NewSession(hostname, c.auth)
	if err != nil {
		return false, err
	}
	defer jnpr.Close()

	// Attempt to apply the provided config without committing.
	err = jnpr.Config(config, "text", false)
	if err != nil {
		return false, err
	}

	diff, err := jnpr.Diff(0)
	if err != nil {
		return false, err
	}

	fmt.Println("DIFF:")
	fmt.Println(diff)

	// TODO: check diff content.
	return true, nil
}
