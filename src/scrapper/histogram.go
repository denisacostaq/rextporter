package scrapper

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/client"
	"github.com/simelo/rextporter/src/util"
)

// histogramClientOptions is a type alias to hold the buckets inside a histogram
type histogramClientOptions []float64

// Histogram implements the Client interface(is able to get histogram metrics through `GetMetric`)
type Histogram struct {
	baseAPIScrapper
	buckets histogramClientOptions
}

func newHistogram(cf client.Factory, parser BodyParser, datasource, jobName, instanceName, nodePath string, buckets histogramClientOptions) Scrapper {
	return Histogram{
		baseAPIScrapper: baseAPIScrapper{
			baseScrapper: baseScrapper{
				jobName:      jobName,
				instanceName: instanceName,
			},
			clientFactory: cf,
			dataSource:    datasource,
			parser:        parser,
			jsonPath:      nodePath,
		},
		buckets: buckets,
	}
}

// GetMetric return a histogram metrics val
func (h Histogram) GetMetric(metricsCollector chan<- prometheus.Metric) (val interface{}, err error) {
	const generalScopeErr = "error scrapping histogram metric"
	var iBody interface{}
	if iBody, err = getData(h.clientFactory, h.parser, metricsCollector); err != nil {
		errCause := "histogram client can not decode the body"
		return val, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var iVal interface{}
	if iVal, err = h.parser.pathLookup(h.jsonPath, iBody); err != nil {
		errCause := fmt.Sprintln("can not get node: ", err.Error())
		return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	histogram, err := createHistogramValueFromData(h.buckets, iVal)
	if err != nil {
		errCause := fmt.Sprintf("can not create histogram value from data %+v.\n%s\n", iVal, err.Error())
		return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	val = histogram
	return val, err
}

// HistogramValue hold the required values to create a histogram metric, the Count, Sum and buckets.
type HistogramValue struct {
	Count   uint64
	Sum     float64
	Buckets map[float64]uint64
}

func newHistogramValue(buckets []float64) HistogramValue {
	val := HistogramValue{
		Count:   0,
		Sum:     0,
		Buckets: make(map[float64]uint64, len(buckets)),
	}
	for _, bucket := range buckets {
		val.Buckets[bucket] = 0
	}
	return val
}

func createHistogramValueFromData(buckets []float64, data interface{}) (histogram HistogramValue, err error) {
	generalScopeErr := "creating histogram from data"
	collection, okCollection := data.([]interface{})
	if !okCollection {
		errCause := fmt.Sprintln("can not assert the metric type as slice")
		return histogram, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	histogram = newHistogramValue(buckets)
	for _, item := range collection {
		histogram.Count++
		metricVal, okMetricVal := item.(float64)
		if !okMetricVal {
			errCause := fmt.Sprintf("can not assert the metric value %+v to type float", item)
			return histogram, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		histogram.Sum += metricVal
		for _, bucket := range buckets {
			if metricVal <= bucket {
				histogram.Buckets[bucket]++
			}
		}
	}
	return histogram, err
}
