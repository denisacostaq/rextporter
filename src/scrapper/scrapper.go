package scrapper

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/client"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// Scrapper get metrics from raw data
type Scrapper interface {
	// GetMetric recive the metrics collector channel and should return the metric val
	GetMetric(metricsCollector chan<- prometheus.Metric) (val interface{}, err error)
	GetJobName() string
	GetInstanceName() string
	GetDataSource() string
}

// FordwaderScrapper get metrics from an already metrics endpoint
type FordwaderScrapper interface {
	// GetMetric should return the metrics vals as raw string
	GetMetric() (val interface{}, err error)
	GetJobName() string
	GetInstanceName() string
}

type baseScrapper struct {
	jobName      string
	instanceName string
}

type baseAPIScrapper struct {
	baseScrapper
	dataSource    string
	clientFactory client.Factory
	parser        BodyParser
	jsonPath      string
}

type baseFordwaderScrapper struct {
	baseScrapper
	clientFactory client.FordwaderFactory
}

func (s baseScrapper) GetJobName() string {
	return s.jobName
}

func (s baseScrapper) GetInstanceName() string {
	return s.instanceName
}

func (s baseAPIScrapper) GetDataSource() string {
	return s.dataSource
}

// BodyParser decode body from different formats, an get some data node
type BodyParser interface {
	decodeBody(body []byte) (val interface{}, err error)
	pathLookup(path string, val interface{}) (node interface{}, err error)
}

// NewScrapper will put all the required info to scrap metrics from the body returned by the client.
func NewScrapper(cf client.Factory, parser BodyParser, resConf config.RextResourceDef, srvConf config.RextServiceDef, mtrConf config.RextMetricDef, nSolver config.RextNodeSolver) (scrapper Scrapper, err error) {
	dataSource := strings.TrimPrefix(resConf.GetResourcePATH(srvConf.GetBasePath()), srvConf.GetBasePath())
	srvOpts := srvConf.GetOptions()
	jobName, err := srvOpts.GetString(config.OptKeyRextServiceDefJobName)
	if err != nil {
		log.WithError(err).Errorln("Can not find jobName")
		return scrapper, err
	}
	instanceName, err := srvOpts.GetString(config.OptKeyRextServiceDefInstanceName)
	if err != nil {
		log.WithError(err).Errorln("Can not find instanceName")
		return scrapper, err
	}
	if len(mtrConf.GetLabels()) > 0 {
		return createVecScrapper(cf, parser, jobName, instanceName, dataSource, nSolver, mtrConf)
	}
	return createAtomicScrapper(cf, parser, jobName, instanceName, dataSource, mtrConf, nSolver)
}

func createVecScrapper(cf client.Factory, parser BodyParser, jobName, instanceName, dataSource string, nSolver config.RextNodeSolver, mtrConf config.RextMetricDef) (scrapper Scrapper, err error) {
	if mtrConf.GetMetricType() == config.KeyMetricTypeCounter || mtrConf.GetMetricType() == config.KeyMetricTypeGauge {
		return newNumericVec(cf, parser, jobName, instanceName, dataSource, nSolver, mtrConf), nil
	}
	log.WithError(errors.New("histogram vec and summary vec are not supported yet")).Errorln("invalid operation")
	return NumericVec{}, config.ErrKeyNotSupported
}

func createAtomicScrapper(cf client.Factory, parser BodyParser, jobName, instanceName, dataSource string, mtrConf config.RextMetricDef, nSolver config.RextNodeSolver) (scrapper Scrapper, err error) {
	if mtrConf.GetMetricType() == config.KeyMetricTypeSummary {
		log.WithError(errors.New("summary scrapper is not supported yet")).Errorln("invalid operation")
		return Histogram{}, config.ErrKeyNotSupported
	}
	if mtrConf.GetMetricType() == config.KeyMetricTypeHistogram {
		bObj, err := mtrConf.GetOptions().GetObject(config.OptKeyRextMetricDefHMetricBuckets)
		if err != nil {
			log.WithError(err).Errorln("no buckets definitions found")
			return scrapper, err
		}
		buckets, okBuckets := bObj.([]float64)
		if !okBuckets {
			log.WithField("val", bObj).Errorln("value is not a float64 array(buckets)")
			return scrapper, config.ErrKeyInvalidType
		}
		return newHistogram(cf, parser, jobName, instanceName, dataSource, nSolver.GetNodePath(), buckets), nil
	}
	return newNumeric(cf, parser, nSolver.GetNodePath(), jobName, instanceName, dataSource), nil
}

func getData(cf client.Factory, p BodyParser, metricsCollector chan<- prometheus.Metric) (data interface{}, err error) {
	const generalScopeErr = "error getting data"
	var cl client.Client
	if cl, err = cf.CreateClient(); err != nil {
		errCause := "can ot create client"
		return data, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var body []byte
	if body, err = cl.GetData(metricsCollector); err != nil {
		errCause := "client can not get data"
		return data, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	if data, err = p.decodeBody(body); err != nil {
		errCause := "scrapper can not decode the body"
		return data, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	return data, err
}
