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

// Handler is the HTTP handler for /check
type Handler struct {
	projectID string
	netconf   internal.NetconfClient
}

// NewHandler returns a Handler with the specified configuration.
func NewHandler(projectID string, netconf internal.NetconfClient) *Handler {
	return &Handler{
		projectID: projectID,
		netconf:   netconf,
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
		writeError(w, err, http.StatusBadRequest)
		return
	}

	config := Config{
		ProjectID: h.projectID,
		Netconf:   h.netconf,
		Provider:  provider,
	}

	registry := prometheus.NewRegistry()
	collector := New(target, config)
	registry.MustRegister(collector)

	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	promHandler.ServeHTTP(w, r)
}

// getProviderForConfig initializes a content.Provider for the specified site.
func (h *Handler) getProviderForConfig(site string) (content.Provider, error) {
	url, err := url.Parse(
		fmt.Sprintf("gs://switch-config-%s/configs/latest/%s.conf",
			h.projectID, site),
	)
	if err != nil {
		return nil, err
	}

	provider, err := content.FromURL(context.Background(), url)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// getSite returns the site name from a FQDN like
// s1.<site>.measurement-lab.org.
func getSite(hostname string) (string, error) {
	re := regexp.MustCompile(`s1\.([a-zA-Z]{3}[0-9]{2}).*`)
	res := re.FindStringSubmatch(hostname)
	if len(res) != 2 {
		return "", fmt.Errorf("cannot extract site from hostname: %s",
			hostname)
	}

	return res[0], nil
}

// writeError writes an error on the provided ResponseWriter and logs it.
func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
	log.WithError(err).Error("Error while processing request")
}
