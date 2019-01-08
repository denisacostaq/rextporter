package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/exporter"
	"github.com/simelo/rextporter/src/tomlconfig"
	"github.com/simelo/rextporter/test/util"
	"github.com/simelo/rextporter/test/util/testrand"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SkycoinSuit struct {
	suite.Suite
	require            *require.Assertions
	rextporterEndpoint string
	rextporterServer   *http.Server
}

func (suite SkycoinSuit) rootConf(fakeNodePort uint16) tomlconfig.RootConfig {
	root := tomlconfig.RootConfig{}
	mtr1 := tomlconfig.Metric{
		Name:             "burnFactor",
		Path:             "/connections/unconfirmed_verify_transaction/burn_factor",
		Options:          tomlconfig.MetricOptions{Type: config.KeyMetricTypeHistogram, Description: "This is a basic description"},
		HistogramOptions: tomlconfig.HistogramOptions{Buckets: []float64{1, 2, 3}},
	}
	mtr2 := tomlconfig.Metric{
		Name: "seq",
		Path: "/blockchain/head/seq",
		// NodeSolverType: "ns0132",
		Options: tomlconfig.MetricOptions{Type: config.KeyMetricTypeGauge, Description: "This is a basic description"},
	}
	mtr3 := tomlconfig.Metric{
		Name: "burnFactorVec",
		Path: "/connections/unconfirmed_verify_transaction/burn_factor",
		Options: tomlconfig.MetricOptions{
			Type:        config.KeyMetricTypeGauge,
			Description: "This is a basic description",
			Labels:      []tomlconfig.Label{tomlconfig.Label{Name: "address", Path: "/connections/address"}},
		},
		HistogramOptions: tomlconfig.HistogramOptions{Buckets: []float64{1, 2, 3}},
	}
	res1 := tomlconfig.ResourcePath{
		Name:           "connections",
		Path:           "/api/v1/network/connections",
		PathType:       "rest_api",
		NodeSolverType: "jsonPath",
		MetricNames:    []string{mtr1.Name, mtr3.Name},
	}
	res2 := tomlconfig.ResourcePath{
		Name:     "Fordwader",
		Path:     "/metrics2",
		PathType: "metrics_fordwader",
	}
	res3 := tomlconfig.ResourcePath{
		Name:           "health",
		Path:           "/api/v1/health",
		PathType:       "rest_api",
		NodeSolverType: "jsonPath",
		MetricNames:    []string{mtr2.Name},
	}
	srv1 := tomlconfig.Service{
		Name:                 "MySuperServer",
		Protocol:             "http",
		Port:                 fakeNodePort,
		BasePath:             "",
		AuthType:             "CSRF",
		TokenHeaderKey:       "X-CSRF-Token",
		GenTokenEndpoint:     "/api/v1/csrf",
		TokenKeyFromEndpoint: "csrf_token",
		Location:             tomlconfig.Server{Location: "localhost"},
		Metrics:              []tomlconfig.Metric{mtr1, mtr2, mtr3},
		ResourcePaths:        []tomlconfig.ResourcePath{res1, res2, res3},
	}
	root.Services = []tomlconfig.Service{srv1}
	return root
}

func (suite *SkycoinSuit) SetupSuite() {
	suite.require = require.New(suite.T())
	mainConfigDir := testrand.RFolderPath()
	err := createDirectoriesWithFullDepth([]string{mainConfigDir})
	suite.Nil(err)
	mainConfFilePath := filepath.Join(mainConfigDir, testrand.RName())
	fakeNodePort, err := readListenPortFromFile()
	suite.Nil(err)
	err = createFullConfig(mainConfFilePath, suite.rootConf(fakeNodePort))
	suite.require.Nil(err)
	conf, err := getConfig(mainConfFilePath)
	suite.require.Nil(err)
	suite.require.False(conf.Validate())
	listenPort := testrand.RandomPort()
	suite.rextporterEndpoint = fmt.Sprintf("http://localhost:%d%s", listenPort, "/metdddrics2")
	suite.rextporterServer = exporter.MustExportMetrics("", "/metdddrics2", listenPort, conf)
	suite.require.NotNil(suite.rextporterServer)
	// NOTE(denisacostaq@gmail.com): Wait for server starts
	time.Sleep(time.Second * 2)
}

func (suite *SkycoinSuit) TearDownSuite() {
	log.Info("Shutting down server...")
	suite.Nil(suite.rextporterServer.Shutdown(context.Context(nil)))
}

func TestSkycoinSuit(t *testing.T) {
	suite.Run(t, new(SkycoinSuit))
}

func (suite *SkycoinSuit) TestDefaultMetricsArePresent() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	mtrs := []string{
		"scrape_duration_seconds",
		"scrape_samples_scraped",
		"data_source_response_duration_seconds",
		"data_source_scrape_duration_seconds",
		"data_source_scrape_samples_scraped",
		"fordwader_response_duration_seconds",
		"fordwader_scrape_duration_seconds"}
	for _, mtr := range mtrs {
		var found bool
		found, err = util.FoundMetric(respBody, mtr)
		suite.Nil(err)
		suite.True(found)
	}
	var found bool
	found, err = util.FoundMetric(respBody, "fordwader_scrape_duration_secondss")
	suite.Nil(err)
	suite.False(found)
}

func (suite *SkycoinSuit) TestFordwadedMetricIsPresent() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var found bool
	found, err = util.FoundMetric(respBody, "go_memstats_mallocs_total1a18ac9b29c6")
	suite.Nil(err)
	suite.True(found)
}

func (suite *SkycoinSuit) TestFordwadedDuplicateMetricInLabeling() {
	// NOTE(denisacostaq@gmail.com): Giving
	// NOTE(denisacostaq@gmail.com): go_goroutines is very useful, this allow automatically check
	// if labeling is working ok, because making go_goroutines exist two times(one with labels, fordwader
	// and other without labels, rextporter) make the parser(expfmt.TextParser) fail

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var found bool
	found, err = util.FoundMetric(respBody, "go_goroutines")
	suite.Nil(err)
	suite.True(found)
}

func (suite *SkycoinSuit) TestConfiguredMetricIsPresent() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var found bool
	found, err = util.FoundMetric(respBody, "seq")
	suite.Nil(err)
	suite.True(found)
}

func (suite *SkycoinSuit) TestConfiguredGaugeMetricValue() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var val float64
	val, err = util.GetGaugeValue(respBody, "seq")
	suite.Nil(err)
	suite.Equal(float64(58894), val)
}

func (suite *SkycoinSuit) TestConfiguredMetricIsNotPresentBecauseIsNotUnderTheRightEndpoint() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var found bool
	found, err = util.FoundMetric(respBody, "burnFactor2")
	suite.Nil(err)
	suite.False(found)
}

func (suite *SkycoinSuit) TestConfiguredHistogramMetric() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var val util.HistogramValue
	val, err = util.GetHistogramValue(respBody, "burnFactor")
	suite.Nil(err)
	suite.Equal(uint64(3), val.SampleCount)
	suite.Equal(float64(2), val.SampleSum)
	suite.Equal(uint64(2), val.Buckets[1])
	suite.Equal(uint64(3), val.Buckets[2])
	suite.Equal(uint64(3), val.Buckets[3])
}

func (suite *SkycoinSuit) TestConfiguredGaugeVecMetric() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.NotNil(respBody)
	var val util.NumericVec
	val, err = util.GetNumericVecValues(respBody, "burnFactorVec")
	suite.Nil(err)
	matchValueForLabels := func(key, val string, number float64, values util.NumericVec) bool {
		for _, value := range values.Values {
			for _, label := range value.Labels {
				if label.Name == key && label.Value == val {
					if value.Number == number {
						return true
					}
					log.WithFields(
						log.Fields{
							"name":            key,
							"value":           val,
							"number":          value.Number,
							"expected_number": number}).Errorln("invalid number value")
					return false
				}
			}
		}
		log.WithFields(
			log.Fields{
				"name":  key,
				"value": val}).Errorln("can not find metric with label")
		return false
	}
	suite.True(matchValueForLabels("address", "139.162.161.41:20002", 2, val))
	suite.True(matchValueForLabels("address", "176.9.84.75:6000", 0, val))
	suite.True(matchValueForLabels("address", "185.120.34.60:6000", 0, val))
}
