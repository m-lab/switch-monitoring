package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/scottdware/go-junos"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/go/siteinfo"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/m-lab/switch-monitoring/internal/netconf"
)

const (
	defaultProjectID = "mlab-sandbox"
	switchHostFormat = "s1.%s.measurement-lab.org"
	siteinfoVersion  = "v1"

	httpClientTimeout = time.Second * 15
)

var (
	flagProject = flag.String("project", defaultProjectID,
		"Use a specific GCP Project ID.")
	flagPrivateKey = flag.String("key", "",
		"Path to the SSH private key to use.")
	flagPassphrase = flag.String("pass", "",
		"Passphrase to decrypt the private key. Can be omitted.")
	flagDebug = flag.Bool("debug", true, "Show debug messages.")

	osExit     = os.Exit
	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf.New(auth)
	}

	httpClient = func(timeout time.Duration) internal.HTTPProvider {
		return &http.Client{
			Timeout: timeout,
		}
	}
)

func main() {
	flag.Parse()

	if *flagDebug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetHandler(text.New(os.Stdout))

	rtx.Must(flagx.ArgsFromEnv(flag.CommandLine), "Cannot parse env args")

	// A private key must be provided.
	if *flagPrivateKey == "" {
		log.Error("The SSH private key must be provided.")
		osExit(1)
	}

	// Initialize Siteinfo provider and the NETCONF client.
	auth := &junos.AuthMethod{
		Username:   "root",
		PrivateKey: *flagPrivateKey,
		Passphrase: *flagPassphrase,
	}
	c := newNetconf(auth)

	// Get switches list.
	log.Infof("Fetching switch list for project %s", *flagProject)
	_, err := switches(*flagProject)
	rtx.Must(err, "Cannot fetch the switch list")

	// TODO: loop over the switches list.
	// This is just an example of the intended usage.
	hash, err := c.GetConfig("s1.lga0t.measurement-lab.org")
	if err != nil {
		log.WithFields(log.Fields{
			"hostname": "s1.lga0t.measurement-lab.org",
		}).WithError(err).Error("Connection failed")
	}

	log.Info(hash)
}

// switches downloads the switches.json file from siteinfo and generates a
// list of valid switch hostnames.
func switches(projectID string) ([]string, error) {
	client := siteinfo.New(projectID, "v1", httpClient(httpClientTimeout))
	switches, err := client.Switches()
	if err != nil {
		return nil, err
	}

	if len(switches) == 0 {
		return nil, fmt.Errorf("the retrieved switches list is empty")
	}

	hosts := make([]string, len(switches))

	i := 0
	for k := range switches {
		hosts[i] = fmt.Sprintf(switchHostFormat, k)
		i++
	}

	return hosts, nil
}
