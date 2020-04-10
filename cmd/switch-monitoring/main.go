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
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/rtx"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/m-lab/switch-monitoring/internal/collector"
	"github.com/m-lab/switch-monitoring/internal/netconf"
)

const (
	defaultListenAddr = ":8080"
	defaultProjectID  = "mlab-sandbox"

	// TODO: use v2 hostnames once they are available.
	// (https://github.com/m-lab/siteinfo/issues/134)
	switchHostFormat  = "s1.%s.measurement-lab.org"
	siteinfoVersion   = "v1"
	httpClientTimeout = time.Second * 15

	// The default cache capacity has been chosen based on the current amount
	// of switches on the platform, plus some significant headroom for future
	// expansion.
	defaultCacheCapacity = 250
	defaultCacheTTL      = 24 * time.Hour
)

var (
	listenAddr = flag.String("listenaddr", defaultListenAddr, "Address to listen on")
	project    = flag.String("project", defaultProjectID,
		"Use a specific GCP Project ID.")

	sshKey = flag.String("ssh.key", "",
		"Path to the SSH private key to use.")
	sshPassphrase = flag.String("ssh.passphrase", "",
		"Passphrase to decrypt the private key. Can be omitted.")

	cacheCapacity = flag.Int("collector.cache-capacity", defaultCacheCapacity,
		"Maximum # of cached responses for the /check endpoint")
	cacheTTL = flag.Duration("collector.cache-ttl", defaultCacheTTL,
		"TTL of cached responses for the /check endpoint")

	debug = flag.Bool("debug", true, "Show debug messages.")

	// Context for the whole program.
	ctx, cancel = context.WithCancel(context.Background())

	osExit = os.Exit

	newNetconf = func(auth *junos.AuthMethod) internal.NetconfClient {
		return netconf.New(auth)
	}
)

func main() {
	var collectorHandler http.Handler

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetHandler(text.New(os.Stdout))

	rtx.Must(flagx.ArgsFromEnv(flag.CommandLine), "Cannot parse env args")

	// A private key must be provided.
	if *sshKey == "" {
		log.Error("The SSH private key must be provided.")
		osExit(1)
	}

	// Initialize Siteinfo provider and the NETCONF client.
	auth := &junos.AuthMethod{
		Username:   "root",
		PrivateKey: *sshKey,
		Passphrase: *sshPassphrase,
	}
	netconf := newNetconf(auth)

	collectorHandler = collector.NewHandler(*flagProject, netconf)

	// Create an in-memory cache to avoid connecting to a switch too often.
	//
	// Assuming we have enough capacity to keep all the responses in the cache,
	// the choice of eviction algorithm does not really make a difference.
	//
	// If we don't have enough capacity to keep all the responses in memory:
	//
	//   - By using MRU, the first capacity - 1 responses will be re-used for
	//     24 hours and switches - capacity + 1 new SSH connections will be
	//     made.
	//   - By using LRU, all the SSH connections will be made each time.
	//
	// Both conditions are very undesirable and we should make sure
	// capacity > switches at all times. However, using MRU significantly
	// limits the impact when this does not hold true anymore.

	memcache, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.MRU),
		memory.AdapterWithCapacity(*cacheCapacity),
	)
	rtx.Must(err, "Cannot initialize in-memory cache.")

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcache),
		cache.ClientWithTTL(*cacheTTL),
	)
	rtx.Must(err, "Cannot initialize in-memory cache client.")

	collectorHandler = cacheClient.Middleware(collectorHandler)

	mux := http.NewServeMux()
	mux.Handle("/v1/check", collectorHandler)

	s := makeHTTPServer(mux)

	rtx.Must(httpx.ListenAndServeAsync(s), "Could not start HTTP server")
	defer s.Close()

	// Keep serving until the context is canceled.
	<-ctx.Done()
}

func makeHTTPServer(h http.Handler) *http.Server {
	return &http.Server{
		Addr:    *listenAddr,
		Handler: h,
	}
}
