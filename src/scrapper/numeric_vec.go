package scrapper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/client"
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// NumericVec implements the Client interface(is able to get numeric metrics through `GetMetric` like Gauge and Counter)
type NumericVec struct {
	baseAPIScrapper
	labels []config.RextLabelDef
}

func newNumericVec(cf client.Factory, p BodyParser, jobName, instanceName, dataSource string, nSolver config.RextNodeSolver, mtrConf config.RextMetricDef) Scrapper {
	return NumericVec{
		baseAPIScrapper: baseAPIScrapper{
			baseScrapper: baseScrapper{
				jobName:      jobName,
				instanceName: instanceName,
			},
			clientFactory: cf,
			dataSource:    dataSource,
			parser:        p,
			jsonPath:      nSolver.GetNodePath(),
		},
		labels: mtrConf.GetLabels(),
	}
}

// NumericVecItemVal can instances a numeric(Gauge or Counter) vec item with the required labels values
type NumericVecItemVal struct {
	Val    float64
	Labels []string
}

// NumericVecVals can instances a numeric(Gauge or Counter) vec with values and corresponding labels
type NumericVecVals []NumericVecItemVal

// GetMetric returns a numeric(Gauge or Counter) vector metric by using remote data.
func (nv NumericVec) GetMetric(metricsCollector chan<- prometheus.Metric) (val interface{}, err error) {
	var iBody interface{}
	if iBody, err = getData(nv.clientFactory, nv.parser, metricsCollector); err != nil {
		log.WithError(err).Errorln("can not get data for numeric vec")
		return val, config.ErrKeyNotSuccessResponse
	}
	var iValColl interface{}
	if iValColl, err = nv.parser.pathLookup(nv.jsonPath, iBody); err != nil {
		log.WithFields(log.Fields{"err": err, "body": iBody, "path": nv.jsonPath}).Errorln("can not get node from body")
		return val, config.ErrKeyDecodingFile
	}
	metricCollection, okMetricCollection := iValColl.([]interface{})
	if !okMetricCollection {
		log.WithField("val", iValColl).Errorln("can not assert value as []interface{}")
		return val, config.ErrKeyInvalidType
	}
	metricsVal := make(NumericVecVals, len(metricCollection))
	for idxIMetricVal, iMetricVal := range metricCollection {
		metricVal, okMetricVal := iMetricVal.(float64)
		if !okMetricVal {
			log.WithField("val", iMetricVal).Errorln("can not assert value as float64")
			return val, config.ErrKeyInvalidType
		}
		metricsVal[idxIMetricVal].Val = metricVal
		metricsVal[idxIMetricVal].Labels = make([]string, len(nv.labels))
		for idxLabel, label := range nv.labels {
			var iLabelValColl interface{}
			ns := label.GetNodeSolver()
			// FIXME(denisacostaq@gmail.com): This should be optimized, calling pathLookup over iBody multiple times,
			// aditionally one for each metric, should be only one for each label
			if iLabelValColl, err = nv.parser.pathLookup(ns.GetNodePath(), iBody); err != nil {
				log.WithFields(log.Fields{"err": err, "body": iBody, "path": ns.GetNodePath()}).Errorln("can not get node from body")
				return val, config.ErrKeyDecodingFile
			}
			iLabelVals, okLabelVal := iLabelValColl.([]interface{})
			if !okLabelVal {
				log.WithField("val", iLabelValColl).Errorln("can not assert value as []interface{}")
				return val, config.ErrKeyInvalidType
			}
			labelVal, okLabelVal := iLabelVals[idxIMetricVal].(string)
			if !okLabelVal {
				log.WithField("val", iLabelVals[idxIMetricVal]).Errorln("can not assert value as string")
				return val, config.ErrKeyInvalidType
			}
			metricsVal[idxIMetricVal].Labels[idxLabel] = labelVal
		}
	}
	return metricsVal, nil
}
