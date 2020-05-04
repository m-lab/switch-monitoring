package netconf

import (
	"errors"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/scottdware/go-junos"
	"golang.org/x/crypto/ssh"
)

var newSession = junos.NewSessionWithConfig

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
	var config *ssh.ClientConfig

	if len(auth.PrivateKey) == 0 {
		return nil, errors.New("no private key specified")
	}
	config, err := netconf.SSHConfigPubKeyFile(auth.Username, auth.PrivateKey, auth.Passphrase)
	if err != nil {
		return nil, err
	}

	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	// This matches the only two key exchange algorithm we use on our switches.
	config.Config.KeyExchanges = []string{"curve25519-sha256@libssh.org",
		"diffie-hellman-group-exchange-sha256"}

	return newSession(host, config)
}
