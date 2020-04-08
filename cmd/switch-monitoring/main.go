package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/scottdware/go-junos"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/m-lab/switch-monitoring/internal/collector"
	"github.com/m-lab/switch-monitoring/internal/netconf"
)

const (
	defaultListenAddr = ":8080"
	defaultPromPort   = ":9600"
	defaultProjectID  = "mlab-sandbox"

	switchHostFormat  = "s1.%s.measurement-lab.org"
	siteinfoVersion   = "v1"
	httpClientTimeout = time.Second * 15
)

var (
	listenAddr  = flag.String("listenaddr", defaultListenAddr, "Address to listen on")
	flagProject = flag.String("project", defaultProjectID,
		"Use a specific GCP Project ID.")

	flagPrivateKey = flag.String("ssh.key", "",
		"Path to the SSH private key to use.")
	flagPassphrase = flag.String("ssh.passphrase", "",
		"Passphrase to decrypt the private key. Can be omitted.")

	flagUsername = flag.String("auth.username", "", "Username for HTTP basic auth")
	flagPassword = flag.String("auth.password", "", "Password for HTTP basic auth")

	flagDebug = flag.Bool("debug", true, "Show debug messages.")

	// Context for the whole program.
	ctx, cancel = context.WithCancel(context.Background())

	osExit = os.Exit

	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf.New(auth)
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
	netconf := newNetconf(auth)

	collectorHandler := collector.NewHandler(*flagProject, netconf)

	mux := http.NewServeMux()
	mux.Handle("/v1/check", collectorHandler)

	s := makeHTTPServer(mux)

	rtx.Must(httpx.ListenAndServeAsync(s), "Could not start HTTP server")
	defer s.Close()

	// Initialize Prometheus server for monitoring.
	// promServer := prometheusx.MustServeMetrics()
	// defer promServer.Close()

	// Keep serving until the context is canceled.
	<-ctx.Done()
}

func makeHTTPServer(h http.Handler) *http.Server {
	return &http.Server{
		Addr:    *listenAddr,
		Handler: h,
	}
}
