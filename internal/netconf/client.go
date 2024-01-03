package netconf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/scottdware/go-junos"
)

var encPasswordRegex = regexp.MustCompile(`(?m)^.*(encrypted-password|\[edit).*$`)

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

	err = jnpr.Lock()
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	defer jnpr.Unlock()

	err = jnpr.Config(config, "text", false)
	if err != nil {
		return false, err
	}

	diff, err := jnpr.Diff(0)
	if err != nil {
		return false, err
	}

	diff = cleanDiff(diff)

	if diff != "" {
		fmt.Printf("Diff for %s:\n%s\n", hostname, diff)
		return false, nil
	}

	return true, nil
}

func cleanDiff(diff string) string {
	s := encPasswordRegex.ReplaceAllString(diff, "")
	return strings.Trim(s, "\n")
}
