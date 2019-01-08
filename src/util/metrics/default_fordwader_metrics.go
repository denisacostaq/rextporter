package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/config"
)

// DefaultFordwaderMetrics default metrics for metrics fordwader
type DefaultFordwaderMetrics struct {
	FordwaderResponseDuration      *prometheus.GaugeVec
	FordwaderScrapeDurationSeconds *prometheus.GaugeVec
}

// NewDefaultFordwaderMetrics create a new DefaultFordwaderMetrics
func NewDefaultFordwaderMetrics() (fordwaderMetrics *DefaultFordwaderMetrics) {
	var instance4JobLabels = []string{config.KeyLabelJob, config.KeyLabelInstance}
	fordwaderMetrics = &DefaultFordwaderMetrics{
		FordwaderResponseDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fordwader_response_duration_seconds",
				Help: "Elapse time(in seconds) to get a response from a fordwader",
			},
			instance4JobLabels,
		),
		FordwaderScrapeDurationSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fordwader_scrape_duration_seconds",
				Help: "Elapse time(in seconds) to get a response from a fordwader scrapper",
			},
			instance4JobLabels,
		),
	}
	return fordwaderMetrics
}

// MustRegister register default metrics for metrics fordwader in prometheus
func (fordwaderMetrics DefaultFordwaderMetrics) MustRegister() {
	prometheus.MustRegister(fordwaderMetrics.FordwaderResponseDuration)
	prometheus.MustRegister(fordwaderMetrics.FordwaderScrapeDurationSeconds)
}
