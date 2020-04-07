package collector

import (
	"context"

	"github.com/apex/log"
	"github.com/m-lab/go/content"
	"github.com/m-lab/switch-monitoring/internal"
	"github.com/m-lab/switch-monitoring/internal/netconf"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	configNotFoundGCS    = "config_not_found_gcs"
	configNotFoundSwitch = "config_not_found_switch"
	configMismatch       = "config_mismatch"
	configMatches        = "ok"
)

type Config struct {
	ProjectID string
	Provider  content.Provider
	Netconf   internal.NetconfClient
}

type ConfigCheckerCollector struct {
	target string
	config Config
	result *prometheus.Desc
}

func New(target string, config Config) *ConfigCheckerCollector {
	return &ConfigCheckerCollector{
		target: target,
		config: config,
		result: prometheus.NewDesc("switch_monitoring_config_match",
			"Configuration check result for this target",
			[]string{"target", "status"}, nil),
	}
}

func (c *ConfigCheckerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.result
}

func (c *ConfigCheckerCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch the latest config from GCS for this target.
	expected, err := c.config.Provider.Get(context.Background())
	if err != nil {
		log.WithFields(log.Fields{"target": c.target}).WithError(err).Error(
			"Cannot fetch latest config from GCS")
		ch <- prometheus.MustNewConstMetric(c.result, prometheus.GaugeValue, 1,
			c.target, configNotFoundGCS)
		return
	}

	// Fetch the actual config from the switch.
	actual, err := c.config.Netconf.GetConfig(c.target)
	if err != nil {
		log.WithFields(log.Fields{"target": c.target}).WithError(err).Error(
			"Cannot fetch config from the switch")
		ch <- prometheus.MustNewConstMetric(c.result, prometheus.GaugeValue, 1,
			c.target, configNotFoundSwitch)
		return
	}

	// Compare them.
	if !netconf.Compare(string(expected), actual) {
		log.WithFields(log.Fields{"target": c.target}).Warn(
			"Switch configuration is different than the archived one.")
		ch <- prometheus.MustNewConstMetric(c.result, prometheus.GaugeValue, 1,
			c.target, configMismatch)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.result, prometheus.GaugeValue, 1,
		c.target, configMatches)
}
