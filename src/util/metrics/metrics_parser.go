package metrics

import (
	"bufio"
	"bytes"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// AppendLables add the labels you give in the list to the metrics names(metricsNames)
// pressent in metricBody(should be a plain text metrics) and return the new metrics
// body with the labels.
// If metricsNames is empty the labels will be apply to all metrics in metricsBody
func AppendLables(metricsNames []string, metricsBody []byte, labels []*io_prometheus_client.LabelPair) ([]byte, error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metricsBody)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return metricsBody, config.ErrKeyDecodingFile
	}
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)
	encoder := expfmt.NewEncoder(writer, expfmt.FmtText)
	for _, mf := range metricFamilies {
		for idxMetric := range mf.Metric {
			if len(metricsNames) == 0 || util.StrSliceContains(metricsNames, *mf.Name) {
				mf.Metric[idxMetric].Label = append(mf.Metric[idxMetric].Label, labels...)
			}
		}
		if err := encoder.Encode(mf); err != nil {
			log.WithFields(log.Fields{"err": err, "metric_family": mf}).Errorln("can not encode metric family")
			return metricsBody, err
		}
	}
	writer.Flush()
	return buff.Bytes(), nil
}

// FindMetricsNamesWithoutLabels return all the metric names that not have at least one of the labels in parameters
func FindMetricsNamesWithoutLabels(metricsBody []byte, labels []string) (mtrNames []string, err error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metricsBody)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return mtrNames, config.ErrKeyDecodingFile
	}
	haveThisLabel := func(labels []*io_prometheus_client.LabelPair, name string) bool {
		for _, label := range labels {
			if name == *label.Name {
				return true
			}
		}
		return false
	}
	for _, mf := range metricFamilies {
		labelFound := make(map[string]bool)
		for idxMetric := range mf.Metric {
			for _, label := range labels {
				labelFound[label] = haveThisLabel(mf.Metric[idxMetric].Label, label)
			}
		}
		var haveAllLabels = true
		for _, label := range labels {
			if v, ok := labelFound[label]; !ok || !v {
				haveAllLabels = false
				break
			}
		}
		if !haveAllLabels {
			mtrNames = append(mtrNames, *mf.Name)
		}
	}
	return mtrNames, nil
}

// FindMetricsNames return all the metric names
func FindMetricsNames(metricsBody []byte) (mtrNames []string, err error) {
	var parser expfmt.TextParser
	in := bytes.NewReader(metricsBody)
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		log.WithError(err).Errorln("error, reading text format failed")
		return mtrNames, config.ErrKeyDecodingFile
	}
	for _, mf := range metricFamilies {
		mtrNames = append(mtrNames, *mf.Name)
	}
	return mtrNames, nil
}
