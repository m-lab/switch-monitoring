package netconf

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/scottdware/go-junos"
)

var (
	connect = func(h string, auth *junos.AuthMethod) (connection, error) {
		return junos.NewSession(h, auth)
	}
)

type connection interface {
	GetConfig(string, ...string) (string, error)
}

// Client is a client to get the switch configuration using the
// NETCONF protocol.
type Client struct {
	auth *junos.AuthMethod
}

// New returns a new NetconfClient.
func New(auth *junos.AuthMethod) Client {
	return Client{
		auth: auth,
	}
}

// GetConfigHash connects to a switch, gets one or more sections via NETCONF,
// removes any comment lines at the beginning, trims whitespace and the
// beginning/end, replaces any encrypted password with "dummy" and returns the
// SHA256 hash of what is left.
//
// The section can be an empty string. In that case, the whole configuration
// will be read.
func (c Client) GetConfigHash(hostname string, section ...string) (string, error) {
	jnpr, err := connect(hostname, c.auth)
	if err != nil {
		return "", err
	}

	config, err := jnpr.GetConfig("text", section...)
	if err != nil {
		return "", err
	}

	// Remove comments (lines starting with '#') and trim the result.
	re := regexp.MustCompile("(?m)^#.*$")
	config = strings.TrimSpace(re.ReplaceAllString(config, ""))

	// Replace all password fields with "dummy".
	// TODO: once we start pre-configuring the switch with a random password,
	// we can remove this step so that the actual passwords are compared.
	re = regexp.MustCompile("encrypted-password.+")
	config = re.ReplaceAllString(config, "encrypted-password \"dummy\";")

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(config)))
	return hash, nil
}
