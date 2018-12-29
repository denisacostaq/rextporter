package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/simelo/rextporter/src/exporter"
	"github.com/simelo/rextporter/src/tomlconfig"
	"github.com/simelo/rextporter/test/integration/testrand"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HealthSuit struct {
	suite.Suite
	require            *require.Assertions
	rextporterEndpoint string
	rextporterServer   *http.Server
}

func (suite HealthSuit) rootConf(fakeNodePort uint16) tomlconfig.RootConfig {
	root := tomlconfig.RootConfig{}
	mtr1 := tomlconfig.Metric{
		Name: "burnFactor",
		Path: "/connections/unconfirmed_verify_transaction/burn_factor",
		// NodeSolverType: "ns0132",
		Options: tomlconfig.MetricOptions{Type: "Gauge", Description: "This is a basic description"},
	}
	mtr2 := tomlconfig.Metric{
		Name: "seq",
		Path: "/blockchain/head/seq",
		// NodeSolverType: "ns0132",
		Options: tomlconfig.MetricOptions{Type: "Gauge", Description: "This is a basic description"},
	}
	res1 := tomlconfig.ResourcePath{
		Name:           "connections",
		Path:           "/api/v1/network/connections",
		PathType:       "rest_api",
		NodeSolverType: "jsonPath",
		MetricNames:    []string{mtr1.Name},
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
		Metrics:              []tomlconfig.Metric{mtr1, mtr2},
		ResourcePaths:        []tomlconfig.ResourcePath{res1, res2, res3},
	}
	root.Services = []tomlconfig.Service{srv1}
	return root
}

func (suite *HealthSuit) SetupSuite() {
	suite.require = require.New(suite.T())
	mainConfigDir := testrand.RFolderPath()
	err := createDirectoriesWithFullDepth([]string{mainConfigDir})
	suite.Nil(err)
	mainConfFilePath := filepath.Join(mainConfigDir, testrand.RName())
	fakeNodePort, err := readListenPortFromFile()
	suite.Nil(err)
	err = createFullConfig(mainConfFilePath, suite.rootConf(fakeNodePort))
	suite.require.Nil(err)

	// NOTE(denisacostaq@gmail.com): Giving

	conf, err := getConfig(mainConfFilePath)
	suite.require.Nil(err)
	suite.require.False(conf.Validate())
	listenPort := testrand.RandomPort()
	suite.rextporterEndpoint = fmt.Sprintf("http://localhost:%d%s", listenPort, "/metdddrics2")
	suite.rextporterServer = exporter.MustExportMetrics("/metdddrics2", listenPort, conf)
	suite.require.NotNil(suite.rextporterServer)
	// NOTE(denisacostaq@gmail.com): Wait for server starts
	time.Sleep(time.Second * 2)
}

func (suite *HealthSuit) TearDownSuite() {
	log.Info("Shutting down server...")
	suite.Nil(suite.rextporterServer.Shutdown(context.Context(nil)))
}

func TestSkycoinHealthSuit(t *testing.T) {
	suite.Run(t, new(HealthSuit))
}

func (suite *HealthSuit) TestDefaultMetricsArePresent() {
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
		found, err = findMetric(respBody, mtr)
		suite.Nil(err)
		suite.True(found)
	}
	var found bool
	found, err = findMetric(respBody, "fordwader_scrape_duration_secondss")
	suite.Nil(err)
	suite.False(found)
}

func (suite *HealthSuit) TestFordwadedMetricIsPresent() {
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
	found, err = findMetric(respBody, "go_memstats_mallocs_total1a18ac9b29c6")
	suite.Nil(err)
	suite.True(found)
}

func (suite *HealthSuit) TestConfiguredMetricIsPresent() {
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
	found, err = findMetric(respBody, "seq")
	suite.Nil(err)
	suite.True(found)
}

func (suite *HealthSuit) TestConfiguredMetricValue() {
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
	val, err = getGaugeValue(respBody, "seq")
	suite.Nil(err)
	suite.Equal(float64(58894), val)
}

func (suite *HealthSuit) TestConfiguredMetricIsNotPresentBecauseServerEndpointInaccesible() {
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
	found, err = findMetric(respBody, "burnFactor")
	suite.Nil(err)
	suite.False(found)
}
