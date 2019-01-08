package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

type scrapDurationInInstance map[string]float64
type scrapDurationInJob map[string]scrapDurationInInstance

type dataSource2Value map[string]float64
type instance2D2V map[string]dataSource2Value
type job2I map[string]instance2D2V

func newDataSourceMappedValue() job2I {
	return make(job2I)
}

var instance4JobLabels = []string{config.KeyLabelJob, config.KeyLabelInstance}

func newDefaultMetrics() *defaultMetrics {
	scrapeDurationsDesc := prometheus.NewDesc(
		"scrape_duration_seconds",
		"Elapse time(in seconds) to get a response from a scrapper",
		instance4JobLabels,
		nil,
	)
	scrapeSamplesScrapedDesc := prometheus.NewDesc(
		"scrape_samples_scraped",
		"The number of samples the target exposed",
		instance4JobLabels,
		nil,
	)
	dataSourceResponseDurationDesc := prometheus.NewDesc(
		"data_source_response_duration_seconds",
		"Elapse time(in seconds) to get a response from a dataSource",
		append(instance4JobLabels, "data_source"),
		nil,
	)
	dataSourceScrapeDurationDesc := prometheus.NewDesc(
		"data_source_scrape_duration_seconds",
		"Elapse time(in seconds) to get a response from a dataSource in a scrapper",
		append(instance4JobLabels, "data_source"),
		nil,
	)
	dataSourceScrapeSamplesDesc := prometheus.NewDesc(
		"data_source_scrape_samples_scraped",
		"The number of samples scrapped in dataSource",
		append(instance4JobLabels, "data_source"),
		nil,
	)
	return &defaultMetrics{
		scrapeDurations:                newScrapDuration(),
		scrapeDurationsDesc:            scrapeDurationsDesc,
		scrapeSamples:                  newScrapDuration(),
		scrapeSamplesScrapedDesc:       scrapeSamplesScrapedDesc,
		dataSourceScrapeDuration:       newDataSourceMappedValue(),
		dataSourceScrapeDurationDesc:   dataSourceScrapeDurationDesc,
		dataSourceScrapeSamples:        newDataSourceMappedValue(),
		dataSourceScrapeSamplesDesc:    dataSourceScrapeSamplesDesc,
		dataSourceResponseDurationDesc: dataSourceResponseDurationDesc,
	}
}

func (instances job2I) add(amount float64, jobName, instanceName, dataSourceName string) {
	dataSources, okDataSources := instances[jobName]
	if okDataSources {
		values, okValues := dataSources[instanceName]
		if okValues {
			values[dataSourceName] += amount
		} else {
			values := make(dataSource2Value)
			values[dataSourceName] = amount
			dataSources[instanceName] = values
		}
	} else {
		values := make(dataSource2Value)
		values[dataSourceName] = amount
		dataSources := make(instance2D2V)
		dataSources[instanceName] = values
		instances[jobName] = dataSources
	}
}

func newScrapDuration() scrapDurationInJob {
	return make(scrapDurationInJob)
}

func (sd scrapDurationInJob) addSeconds(amount float64, jobName, instanceName string) {
	instances, okInstances := sd[jobName]
	if okInstances {
		instances[instanceName] += amount
	} else {
		instances := make(scrapDurationInInstance)
		instances[instanceName] = amount
		sd[jobName] = instances
	}
}

type defaultMetrics struct {
	scrapeDurations                scrapDurationInJob
	scrapeDurationsDesc            *prometheus.Desc
	scrapeSamples                  scrapDurationInJob
	scrapeSamplesScrapedDesc       *prometheus.Desc
	dataSourceScrapeDuration       job2I
	dataSourceScrapeDurationDesc   *prometheus.Desc
	dataSourceScrapeSamples        job2I
	dataSourceScrapeSamplesDesc    *prometheus.Desc
	dataSourceResponseDurationDesc *prometheus.Desc
}

func (defMetrics defaultMetrics) describe(ch chan<- *prometheus.Desc) {
	ch <- defMetrics.scrapeDurationsDesc
	ch <- defMetrics.scrapeSamplesScrapedDesc
	ch <- defMetrics.dataSourceScrapeDurationDesc
	ch <- defMetrics.dataSourceScrapeSamplesDesc
	ch <- defMetrics.dataSourceResponseDurationDesc
}

func (defMetrics *defaultMetrics) reset() {
	defMetrics.scrapeDurations = newScrapDuration()
	defMetrics.scrapeSamples = newScrapDuration()
	defMetrics.dataSourceScrapeDuration = newDataSourceMappedValue()
	defMetrics.dataSourceScrapeSamples = newDataSourceMappedValue()
}

func (defMetrics defaultMetrics) collectDefaultMetrics(ch chan<- prometheus.Metric) {
	for jobName, job := range defMetrics.scrapeDurations {
		for instanceName, val := range job {
			labels := []string{jobName, instanceName}
			if metric, err := prometheus.NewConstMetric(defMetrics.scrapeDurationsDesc, prometheus.GaugeValue, val, labels...); err == nil {
				ch <- metric
			} else {
				log.WithError(err).Errorln("collectDefaultMetrics -> scrapeDurationDesc")
			}
		}
	}
	for jobName, job := range defMetrics.scrapeSamples {
		for instanceName, val := range job {
			labels := []string{jobName, instanceName}
			if metric, err := prometheus.NewConstMetric(defMetrics.scrapeSamplesScrapedDesc, prometheus.GaugeValue, val, labels...); err == nil {
				ch <- metric
			} else {
				log.WithError(err).Errorln("collectDefaultMetrics -> scrapeSamplesScrapedDesc")
			}
		}
	}
	for jobName, instances := range defMetrics.dataSourceScrapeDuration {
		for instanceName, values := range instances {
			for dataSourceName, val := range values {
				labels := []string{jobName, instanceName, dataSourceName}
				if metric, err := prometheus.NewConstMetric(defMetrics.dataSourceScrapeDurationDesc, prometheus.GaugeValue, val, labels...); err == nil {
					ch <- metric
				} else {
					log.WithError(err).Errorln("collectDefaultMetrics -> dataSourceScrapeDuration")
				}
			}
		}
	}
	for jobName, instances := range defMetrics.dataSourceScrapeSamples {
		for instanceName, values := range instances {
			for dataSourceName, val := range values {
				labels := []string{jobName, instanceName, dataSourceName}
				if metric, err := prometheus.NewConstMetric(defMetrics.dataSourceScrapeSamplesDesc, prometheus.GaugeValue, val, labels...); err == nil {
					ch <- metric
				} else {
					log.WithError(err).Errorln("collectDefaultMetrics -> dataSourceScrapeSamples")
				}
			}
		}
	}
}
