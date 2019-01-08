package scrapper

import (
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/simelo/rextporter/src/client"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	"github.com/simelo/rextporter/src/util/metrics"
	mutil "github.com/simelo/rextporter/src/util/metrics"
	log "github.com/sirupsen/logrus"
)

// MetricsForwader is a scrapper kind capable to forward a metrics endpoint with job and instance labels at least
type MetricsForwader struct {
	baseFordwaderScrapper
	defFordwaderMetrics *metrics.DefaultFordwaderMetrics
}

// GetJobName return the name of the job(service)
func (scrapper MetricsForwader) GetJobName() string {
	return scrapper.jobName
}

// GetInstanceName return the name of the instance(ip:port)
func (scrapper MetricsForwader) GetInstanceName() string {
	return scrapper.instanceName
}

// NewMetricsForwader create a scrapper that handle the forwaded metrics
func NewMetricsForwader(pmcls client.ProxyMetricClientCreator, fDefMetrics *metrics.DefaultFordwaderMetrics) FordwaderScrapper {
	return MetricsForwader{
		baseFordwaderScrapper: baseFordwaderScrapper{
			baseScrapper: baseScrapper{
				jobName:      pmcls.JobName,
				instanceName: pmcls.InstanceName,
			},
			clientFactory: pmcls,
		},
		defFordwaderMetrics: fDefMetrics,
	}
}

// GetMetric return the original metrics but with a service name as prefix in his names
func (scrapper MetricsForwader) GetMetric() (val interface{}, err error) {
	getFordwadedMetrics := func() (data []byte, err error) {
		successResponse := false
		defer func(startTime time.Time) {
			duration := time.Since(startTime).Seconds()
			labels := []string{scrapper.GetJobName(), scrapper.GetInstanceName()}
			if successResponse {
				scrapper.defFordwaderMetrics.FordwaderScrapeDurationSeconds.WithLabelValues(labels...).Set(duration)
			}
		}(time.Now().UTC())
		generalScopeErr := "Error getting data for metrics fordwader"
		recorder := httptest.NewRecorder()
		var cl client.FordwaderClient
		if cl, err = scrapper.clientFactory.CreateClient(); err != nil {
			errCause := "can not create client"
			return data, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		var exposedMetricsData []byte
		if exposedMetricsData, err = cl.GetData(); err != nil {
			log.WithError(err).Error("error getting metrics from service " + scrapper.GetJobName())
			errCause := "can not get the data"
			return data, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		job := config.KeyLabelJob
		instance := config.KeyLabelInstance
		prefixed, err := mutil.AppendLables(
			nil,
			exposedMetricsData,
			[]*io_prometheus_client.LabelPair{
				&io_prometheus_client.LabelPair{
					Name:  &job,
					Value: &scrapper.jobName,
				},
				&io_prometheus_client.LabelPair{
					Name:  &instance,
					Value: &scrapper.instanceName,
				},
			},
		)
		if err != nil {
			log.WithError(err).Errorln("Can not append default labels for self metric inside rextporter")
			return nil, config.ErrKeyDecodingFile
		}
		if count, err := recorder.Write(prefixed); err != nil || count != len(prefixed) {
			if err != nil {
				log.WithError(err).Errorln("error writing prefixed content")
			}
			if count != len(prefixed) {
				log.WithFields(log.Fields{
					"wrote":    count,
					"required": len(prefixed),
				}).Errorln("no enough content wrote")
				return nil, errors.New("no enough content wrote")
			}
		}
		if data, err = ioutil.ReadAll(recorder.Body); err != nil {
			log.WithError(err).Errorln("can not read recorded custom data")
			return nil, err
		}
		successResponse = true
		return data, nil
	}
	if fordwadedMetrics, err := getFordwadedMetrics(); err == nil {
		val = fordwadedMetrics
	} else {
		log.WithError(err).Errorln("error getting fordwaded metrics")
	}
	return val, nil
}
