package util

import (
	"bytes"
	"strings"

	"github.com/prometheus/common/expfmt"
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// FoundMetric return true if found a metric with mtrName inside a metrics endpoint response as plain text.
func FoundMetric(metrics []byte, mtrName string) (bool, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metrics)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return false, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		if mtrName == *mf.Name {
			return true, nil
		}
	}
	return false, err
}

// GetGaugeValue return a number(the gauge value) if found a metric with mtrName and if a Gauge kind metric
// from a metrics endpoint response as plain text.
func GetGaugeValue(metrics []byte, mtrName string) (float64, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metrics)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return -1, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		if mtrName == *mf.Name {
			if (*mf.Type).String() != strings.ToUpper(config.KeyMetricTypeGauge) {
				log.WithFields(log.Fields{
					"current_type": (*mf.Type).String(),
					"looking_for":  strings.ToUpper(config.KeyMetricTypeHistogram),
				}).Errorln("metric is not one of the spected type")
				return -1, config.ErrKeyInvalidType
			}
			return *mf.Metric[0].Gauge.Value, nil
		}
	}
	log.WithField("wanted_name", mtrName).Errorln("metric name not found")
	return -1, config.ErrKeyNotFound
}

// GetCounterValue return a number(the counter value) if found a metric with mtrName and if a Counter kind metric
// from a metrics endpoint response as plain text.
func GetCounterValue(metrics []byte, mtrName string) (float64, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metrics)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return -1, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		if mtrName == *mf.Name {
			if (*mf.Type).String() != strings.ToUpper(config.KeyMetricTypeCounter) {
				log.WithFields(log.Fields{
					"current_type": (*mf.Type).String(),
					"looking_for":  strings.ToUpper(config.KeyMetricTypeHistogram),
				}).Errorln("metric is not one of the spected type")
				return -1, config.ErrKeyInvalidType
			}
			return *mf.Metric[0].Counter.Value, nil
		}
	}
	log.WithField("wanted_name", mtrName).Errorln("metric name not found")
	return -1, config.ErrKeyNotFound
}

// HistogramValue have some field to fill a histograms instance an compare against some receive value in
// an integration test
type HistogramValue struct {
	SampleCount uint64
	SampleSum   float64
	Buckets     map[float64]uint64
}

// GetHistogramValue return a HistogramValue instance looking into the metrics body for a
// metric with the name mtrName
func GetHistogramValue(metrics []byte, mtrName string) (HistogramValue, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metrics)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return HistogramValue{}, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		if mtrName == *mf.Name {
			if (*mf.Type).String() != strings.ToUpper(config.KeyMetricTypeHistogram) {
				log.WithFields(log.Fields{
					"current_type": (*mf.Type).String(),
					"looking_for":  strings.ToUpper(config.KeyMetricTypeHistogram),
				}).Errorln("metric is not one of the spected type")
				return HistogramValue{}, config.ErrKeyInvalidType
			}
			mtr := *mf.Metric[0]
			hv := HistogramValue{
				SampleCount: *mtr.Histogram.SampleCount,
				SampleSum:   *mtr.Histogram.SampleSum,
				Buckets:     make(map[float64]uint64),
			}
			for _, b := range mtr.Histogram.Bucket {
				hv.Buckets[*b.UpperBound] = *b.CumulativeCount
			}
			return hv, nil
		}
	}
	log.WithField("wanted_name", mtrName).Errorln("metric name not found")
	return HistogramValue{}, config.ErrKeyNotFound
}

// LabelValue have some field to fill a label instance an compare against some received value in
// an integration test
type LabelValue struct {
	Name  string
	Value string
}

// MetricVal have some field to fill a labeled metric instance an compare against some received value in
// an integration test
type MetricVal struct {
	Labels []LabelValue
	Number float64
}

// NumericVec have some field to fill a metric vec(with labels) instance an compare against some received value in
// an integration test
type NumericVec struct {
	Values []MetricVal
}

// GetNumericVecValues return a NumericVec instance looking into the metrics body for a
// metric with the name mtrName
func GetNumericVecValues(metrics []byte, mtrName string) (NumericVec, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metrics)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return NumericVec{}, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		if mtrName == *mf.Name {
			if (*mf.Type).String() != strings.ToUpper(config.KeyMetricTypeGauge) && (*mf.Type).String() != strings.ToUpper(config.KeyMetricTypeCounter) {
				log.WithFields(log.Fields{
					"current_type": (*mf.Type).String(),
					"looking_for":  strings.ToUpper(config.KeyMetricTypeHistogram),
				}).Errorln("metric is not one of the spected type")
				return NumericVec{}, config.ErrKeyInvalidType
			}
			nv := NumericVec{}
			for _, mtrVal := range mf.Metric {
				var val float64
				if (*mf.Type).String() == strings.ToUpper(config.KeyMetricTypeGauge) {
					val = *((*mtrVal).Gauge.Value)
				} else {
					val = *((*mtrVal).Counter.Value)
				}
				mv := MetricVal{Number: val}
				for _, label := range (*mtrVal).Label {
					lv := LabelValue{Name: *label.Name, Value: *label.Value}
					mv.Labels = append(mv.Labels, lv)
				}
				nv.Values = append(nv.Values, mv)
			}
			return nv, nil
		}
	}
	log.WithField("wanted_name", mtrName).Errorln("metric name not found")
	return NumericVec{}, config.ErrKeyNotFound
}
