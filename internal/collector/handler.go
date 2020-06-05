package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/apex/log"
	"github.com/m-lab/go/content"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var parseURL = url.Parse

// Handler is the HTTP handler for /check
type Handler struct {
	projectID     string
	netconf       internal.NetconfClient
	getConfigFunc func(context.Context, *url.URL) (content.Provider, error)
}

// NewHandler returns a Handler with the specified configuration.
func NewHandler(projectID string, netconf internal.NetconfClient) *Handler {
	return &Handler{
		projectID:     projectID,
		netconf:       netconf,
		getConfigFunc: content.FromURL,
	}
}

// ServeHTTP handles GET requests to the /check endpoint, parsing the target
// parameter and delegating writing the actual response to promhttp.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	target := r.URL.Query().Get("target")
	if len(target) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL parameter 'target' is missing"))
		log.Info("URL parameter 'target' is missing")
		return
	}

	site, err := getSite(target)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	provider, err := h.getProviderForConfig(site)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	config := Config{
		ProjectID: h.projectID,
		Netconf:   h.netconf,
		Provider:  provider,
	}

	// This collector depends on external parameters (target) and only returns
	// one metric. This is not how custom collectors usually work.
	// To make it possible we create a new Collector that only collects metrics
	// for the requested targets every time, and we create a new Registry,
	// register this temporary Collector and let Prometheus write the response.
	//
	// A possible alternative approach that was considered is having a single
	// /metrics endpoint that, when called, checked every switch on the
	// platform and returned metrics for all of them. However, this would easily
	// exceed Prometheus' scraping time, would imply an additional dependency on
	// siteinfo (to fetch the current switches.json) and would not benefit from
	// the randomized scraping interval Prometheus implements.

	registry := prometheus.NewRegistry()
	collector := New(target, config)
	registry.MustRegister(collector)

	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	promHandler.ServeHTTP(w, r)
}

// getProviderForConfig initializes a content.Provider for the specified site.
func (h *Handler) getProviderForConfig(site string) (content.Provider, error) {
	url, err := parseURL(
		fmt.Sprintf("gs://switch-config-%s/configs/current/%s.conf",
			h.projectID, site),
	)
	if err != nil {
		return nil, err
	}

	provider, err := h.getConfigFunc(context.Background(), url)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// getSite returns the site name from a FQDN like
// s1.<site>.measurement-lab.org.
func getSite(hostname string) (string, error) {
	re := regexp.MustCompile(`s1(?:\.|-)([a-z]{3}[0-9ct]{2}).*`)
	res := re.FindStringSubmatch(hostname)
	if len(res) != 2 {
		return "", fmt.Errorf("cannot extract site from hostname: %s",
			hostname)
	}

	return res[1], nil
}

// writeError writes an error on the provided ResponseWriter and logs it.
func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
	log.WithError(err).Error("Error while processing request")
}
