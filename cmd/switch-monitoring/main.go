package main

import (
	"flag"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/scottdware/go-junos"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/switch-monitoring/cmd/switch-monitoring/client"
	"github.com/m-lab/switch-monitoring/cmd/switch-monitoring/siteinfo"
)

const defaultProjectID = "mlab-oti"

var (
	flagProject = flag.String("project", defaultProjectID,
		"Use a specific GCP Project ID.")
	flagPrivateKey = flag.String("key", "",
		"Path to the SSH private key to use.")
	flagPassphrase = flag.String("pass", "",
		"Passphrase to decrypt the private key. Can be omitted.")
	flagDebug = flag.Bool("debug", true, "Show debug messages.")
)

func main() {
	if *flagDebug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetHandler(text.New(os.Stdout))

	flag.Parse()
	rtx.Must(flagx.ArgsFromEnv(flag.CommandLine), "Cannot parse env args")

	// A private key must be provided.
	if *flagPrivateKey == "" {
		log.Error("The SSH private key must be provided.")
		os.Exit(1)
	}

	// Initialize Siteinfo provider and the NETCONF client.
	siteinfo := &siteinfo.Siteinfo{ProjectID: *flagProject}
	auth := &junos.AuthMethod{
		Username:   "root",
		PrivateKey: *flagPrivateKey,
		Passphrase: *flagPassphrase,
	}
	c := client.New(auth)

	// Get switches list.
	log.Infof("Fetching switch list for project %s", *flagProject)
	_, err := siteinfo.Switches()
	rtx.Must(err, "Cannot fetch the switch list")

	// TODO: loop over the switches list.
	hash, err := c.GetConfigHash("s1.lga0t.measurement-lab.org")
	if err != nil {
		log.WithFields(log.Fields{
			"hostname": "",
		}).WithError(err).Error("Connection failed")
	}

	log.Info(hash)

}
