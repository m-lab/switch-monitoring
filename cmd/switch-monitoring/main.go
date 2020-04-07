package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/scottdware/go-junos"

	"github.com/m-lab/go/content"
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

	osExit        = os.Exit
	configFromURL = func(ctx context.Context, u *url.URL) (content.Provider, error) {
		return content.FromURL(ctx, u)
	}

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

	// Get all the switch hostnames.
	log.Infof("Fetching switch hostnames for project %s", *flagProject)
	sites, err := sites(*flagProject)
	rtx.Must(err, "Cannot fetch the switch list")

	// TODO: Here we should start a Prometheus instance and check the configs
	// only when the /metrics endpoint is called, caching as long as we deem
	// necessary (e.g. 24 hours.)
	//
	// For now, the check only happens once.

	checkAll(c, sites)
}

func checkAll(c internal.NetconfClient, sites []string) {
	for _, site := range sites {
		hostname := fmt.Sprintf(switchHostFormat, site)
		conf, err := c.GetConfig(hostname)
		if err != nil {
			// TODO: expose this error as a Prometheus metric.
			log.WithFields(log.Fields{
				"site": site,
			}).WithError(err).Error("Connection failed")
			continue
		}

		// Get the archived config from GCS.
		url, err := url.Parse(fmt.Sprintf(
			"gs://switch-config-%s/configs/current/%s.conf",
			*flagProject, site))
		rtx.Must(err, "Cannot create URL for site %s", site)

		prov, err := configFromURL(context.Background(), url)
		rtx.Must(err, "Cannot create GCS provider for site %", site)

		archived, err := prov.Get(context.Background())
		if err != nil {
			log.WithFields(log.Fields{
				"site": site,
			}).WithError(err).Error("Cannot retrieve archived config")
			continue
		}

		if !netconf.Compare(conf, string(archived)) {
			log.WithFields(log.Fields{
				"site": site,
			}).Warn("Current and archived configuration for %s do not match.")
		}
	}
}

// sites downloads the switches.json file from siteinfo and generates a
// list of sites.
func sites(projectID string) ([]string, error) {
	client := siteinfo.New(projectID, "v1", httpClient(httpClientTimeout))
	switches, err := client.Switches()
	if err != nil {
		return nil, err
	}

	if len(switches) == 0 {
		return nil, fmt.Errorf("the retrieved switches list is empty")
	}

	sitesList := make([]string, len(switches))

	i := 0
	for k := range switches {
		sitesList[i] = k
		i++
	}

	return sitesList, nil
}
