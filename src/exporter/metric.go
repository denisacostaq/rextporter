package exporter

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/cache"
	"github.com/simelo/rextporter/src/client"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/scrapper"
	"github.com/simelo/rextporter/src/util"
	"github.com/simelo/rextporter/src/util/metrics"
	log "github.com/sirupsen/logrus"
)

func createMetricsForwaders(conf config.RextRoot, fDefMetrics *metrics.DefaultFordwaderMetrics) (fordwaderScrappers []scrapper.FordwaderScrapper, err error) {
	generalScopeErr := "can not create metrics Middleware"
	services := conf.GetServices()
	for _, srvConf := range services {
		var metricFordwaderCreator client.ProxyMetricClientCreator
		resources := srvConf.GetResources()
		for _, resConf := range resources {
			if resConf.GetType() == "metrics_fordwader" {
				if metricFordwaderCreator, err = client.CreateProxyMetricClientCreator(resConf, srvConf, fDefMetrics); err != nil {
					errCause := fmt.Sprintln("error creating metric client: ", err.Error())
					return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
				}
				fordwaderScrappers = append(fordwaderScrappers, scrapper.NewMetricsForwader(metricFordwaderCreator, fDefMetrics))
			}
		}
	}
	return fordwaderScrappers, nil
}

// constMetric has a scrapper to get remote data, can be a previously cached content
type constMetric struct {
	kind       string
	scrapper   scrapper.Scrapper
	metricDesc *prometheus.Desc
}

type endpointData2MetricsConsumer map[string][]constMetric

func createMetrics(cache cache.Cache, conf config.RextRoot, dataSourceResponseDurationDesc *prometheus.Desc) (metrics endpointData2MetricsConsumer, err error) {
	generalScopeErr := "can not create metrics"
	metrics = make(endpointData2MetricsConsumer)
	for _, srvConf := range conf.GetServices() {
		for _, resConf := range srvConf.GetResources() {
			k := resConf.GetResourcePATH(srvConf.GetBasePath())
			var m constMetric
			for _, mtrConf := range resConf.GetMetricDefs() {
				nSolver := mtrConf.GetNodeSolver()
				if m, err = createConstMetric(cache, resConf, srvConf, mtrConf, nSolver, dataSourceResponseDurationDesc); err != nil {
					errCause := fmt.Sprintln(fmt.Sprintf("error creating metric client for %s metric of kind %s. ", mtrConf.GetMetricName(), mtrConf.GetMetricType()), err.Error())
					return metrics, util.ErrorFromThisScope(errCause, generalScopeErr)
				}
				metrics[k] = append(metrics[k], m)
			}
		}
	}
	return metrics, err
}

func createConstMetric(cache cache.Cache, resConf config.RextResourceDef, srvConf config.RextServiceDef, mtrConf config.RextMetricDef, nSolver config.RextNodeSolver, dataSourceResponseDurationDesc *prometheus.Desc) (metric constMetric, err error) {
	generalScopeErr := "can not create metric " + mtrConf.GetMetricName()
	if len(mtrConf.GetMetricName()) == 0 {
		log.Errorln("metric name is required")
		return metric, config.ErrKeyEmptyValue
	}
	var ccf client.CacheableFactory
	if ccf, err = client.CreateAPIRestCreator(resConf, srvConf, dataSourceResponseDurationDesc); err != nil {
		errCause := fmt.Sprintln("error creating metric client: ", err.Error())
		return metric, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	cc := client.CatcherCreator{Cache: cache, ClientFactory: ccf}
	var numScrapper scrapper.Scrapper
	if numScrapper, err = scrapper.NewScrapper(cc, scrapper.JSONParser{}, resConf, srvConf, mtrConf, nSolver); err != nil {
		errCause := fmt.Sprintln("error creating metric client: ", err.Error())
		return metric, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var labelsNames []string
	for _, label := range mtrConf.GetLabels() {
		labelsNames = append(labelsNames, label.GetName())
	}
	labels := append(labelsNames, instance4JobLabels...)
	metric = constMetric{
		kind:     mtrConf.GetMetricType(),
		scrapper: numScrapper,
		// FIXME(denisacostaq@gmail.com): if you use a duplicated name can panic?
		metricDesc: prometheus.NewDesc(mtrConf.GetMetricName(), mtrConf.GetMetricDescription(), labels, nil),
	}
	return metric, err
}
