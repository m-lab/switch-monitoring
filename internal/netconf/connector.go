package netconf

import (
	"errors"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/scottdware/go-junos"
	"golang.org/x/crypto/ssh"
)

const defaultTimeout = 15 * time.Second

var newSession = junos.NewSessionWithConfig

// These types provide an abstraction for the underlying connector and
// connection to a NETCONF-enabled device, so that they can be unit tested.

type connection interface {
	GetConfig(string, ...string) (string, error)
	Close()
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

	config.Timeout = defaultTimeout

	// Every time the switch is rebooted, a new host key is generated.
	// Since we don't have any mean to track host key changes at the moment,
	// and we don't know which key is the "correct" one, we do not check the
	// key here.
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	// This matches the only two key exchange algorithm we use on our switches.
	config.Config.KeyExchanges = []string{"curve25519-sha256@libssh.org",
		"diffie-hellman-group-exchange-sha256"}

	return newSession(host, config)
}
