package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/exporter"
	"github.com/simelo/rextporter/src/toml2config"
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

func (suite *SkycoinSuit) SetupSuite() {
	suite.require = require.New(suite.T())
	mainConfFilePath := "tomlconfig/main.toml"
	tomlConf, err := tomlconfig.ReadConfigFromFileSystem(mainConfFilePath)
	suite.Nil(err)
	var conf config.RextRoot
	conf, err = toml2config.Fill(tomlConf)
	suite.Nil(err)
	listenPort := testrand.RandomPort()
	suite.rextporterEndpoint = fmt.Sprintf("http://localhost:%d%s", listenPort, "/metrics")
	suite.rextporterServer = exporter.MustExportMetrics("", "/metrics", listenPort, conf)
	suite.require.NotNil(suite.rextporterServer)
	// NOTE(denisacostaq@gmail.com): Wait for server starts
	time.Sleep(time.Second * 2)
}

func (suite *SkycoinSuit) TearDownSuite() {
	log.Info("Shutting down server...")
	suite.Nil(suite.rextporterServer.Shutdown(context.Context(nil)))
}

func TestSkycoinSuitSuit(t *testing.T) {
	suite.Run(t, new(SkycoinSuit))
}

func (suite *SkycoinSuit) TestMetricsAreExported() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resp, err := http.Get(suite.rextporterEndpoint)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
}

func (suite *SkycoinSuit) TestDefaultMetricsForAPIScrapperArePresent() {
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
	}
	for _, mtr := range mtrs {
		var found bool
		found, err = util.FoundMetric(respBody, mtr)
		suite.Nil(err)
		suite.True(found)
	}
	var found bool
	found, err = util.FoundMetric(respBody, "data_source_scrape_samples_scrapeds")
	suite.Nil(err)
	suite.False(found)
}

func (suite *SkycoinSuit) TestDefaultMetricsForMetricsFordwaderArePresent() {
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
		"fordwader_response_duration_seconds",
		"fordwader_scrape_duration_seconds",
	}
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

func (suite *SkycoinSuit) TestHealthSeq() {
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
	found, err = util.FoundMetric(respBody, "health_seq")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetCounterValue(respBody, "health_seq")
	suite.Nil(err)
	suite.Equal(float64(180), val)
}

func (suite *SkycoinSuit) TestHealthFee() {
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
	found, err = util.FoundMetric(respBody, "health_fee")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_fee")
	suite.Nil(err)
	suite.Equal(float64(2265261), val)
}

func (suite *SkycoinSuit) TestHealthUnspents() {
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
	found, err = util.FoundMetric(respBody, "health_unspents")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_unspents")
	suite.Nil(err)
	suite.Equal(float64(218), val)
}

func (suite *SkycoinSuit) TestHealthUnconfirmed() {
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
	found, err = util.FoundMetric(respBody, "health_unconfirmed")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_unconfirmed")
	suite.Nil(err)
	suite.Equal(float64(1), val)
}

func (suite *SkycoinSuit) TestHealthOpenConnections() {
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
	found, err = util.FoundMetric(respBody, "health_open_connections")
	suite.Nil(err)
	suite.True(found)
	_, err = util.GetGaugeValue(respBody, "health_open_connections")
	suite.Nil(err)
}

func (suite *SkycoinSuit) TestHealthOutgoingConnections() {
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
	found, err = util.FoundMetric(respBody, "health_outgoing_connections")
	suite.Nil(err)
	suite.True(found)
	_, err = util.GetGaugeValue(respBody, "health_outgoing_connections")
	suite.Nil(err)
}

func (suite *SkycoinSuit) TestHealthIncomingConnections() {
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
	found, err = util.FoundMetric(respBody, "health_incoming_connections")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_incoming_connections")
	suite.Nil(err)
	suite.Equal(float64(0), val)
}

func (suite *SkycoinSuit) TestHealthUserVerifyBurnFactor() {
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
	found, err = util.FoundMetric(respBody, "health_user_verify_burn_factor")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_user_verify_burn_factor")
	suite.Nil(err)
	suite.Equal(float64(2), val)
}

func (suite *SkycoinSuit) TestHealthUserVerifyMaxTransactionSize() {
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
	found, err = util.FoundMetric(respBody, "health_user_verify_max_transaction_size")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_user_verify_max_transaction_size")
	suite.Nil(err)
	suite.Equal(float64(32768), val)
}

func (suite *SkycoinSuit) TestHealthUserVerifyMaxDecimals() {
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
	found, err = util.FoundMetric(respBody, "health_user_verify_max_decimals")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_user_verify_max_decimals")
	suite.Nil(err)
	suite.Equal(float64(3), val)
}

func (suite *SkycoinSuit) TestHealthUnconfirmedVerifyBurnFactor() {
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
	found, err = util.FoundMetric(respBody, "health_unconfirmed_verify_burn_factor")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_unconfirmed_verify_burn_factor")
	suite.Nil(err)
	suite.Equal(float64(2), val)
}

func (suite *SkycoinSuit) TestHealthUnconfirmedVerifyMaxTransactionSize() {
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
	found, err = util.FoundMetric(respBody, "health_unconfirmed_verify_max_transaction_size")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_unconfirmed_verify_max_transaction_size")
	suite.Nil(err)
	suite.Equal(float64(32768), val)
}

func (suite *SkycoinSuit) TestHealthUnconfirmedVerifyMaxDecimals() {
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
	found, err = util.FoundMetric(respBody, "health_unconfirmed_verify_max_decimals")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "health_unconfirmed_verify_max_decimals")
	suite.Nil(err)
	suite.Equal(float64(3), val)
}

func (suite *SkycoinSuit) TestBlockchainMetadataSeq() {
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
	found, err = util.FoundMetric(respBody, "blockchain_metadata_seq")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetCounterValue(respBody, "blockchain_metadata_seq")
	suite.Nil(err)
	suite.Equal(float64(180), val)
}

func (suite *SkycoinSuit) TestBlockchainMetadataFee() {
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
	found, err = util.FoundMetric(respBody, "blockchain_metadata_fee")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "blockchain_metadata_fee")
	suite.Nil(err)
	suite.Equal(float64(2265261), val)
}

func (suite *SkycoinSuit) TestBlockchainMetadataUnspents() {
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
	found, err = util.FoundMetric(respBody, "blockchain_metadata_unspents")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "blockchain_metadata_unspents")
	suite.Nil(err)
	suite.Equal(float64(218), val)
}

func (suite *SkycoinSuit) TestBlockchainMetadataUnconfirmed() {
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
	found, err = util.FoundMetric(respBody, "blockchain_metadata_unconfirmed")
	suite.Nil(err)
	suite.True(found)
	var val float64
	val, err = util.GetGaugeValue(respBody, "blockchain_metadata_unconfirmed")
	suite.Nil(err)
	suite.Equal(float64(1), val)
}

func (suite *SkycoinSuit) TestBlockchainProgressCurrent() {
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
	found, err = util.FoundMetric(respBody, "blockchain_progress_current")
	suite.Nil(err)
	suite.True(found)
	// var val float64
	_, err = util.GetCounterValue(respBody, "blockchain_progress_current")
	suite.Nil(err)
}

func (suite *SkycoinSuit) TestBlockchainProgressHighest() {
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
	found, err = util.FoundMetric(respBody, "blockchain_progress_highest")
	suite.Nil(err)
	suite.True(found)
	// var val float64
	_, err = util.GetGaugeValue(respBody, "blockchain_progress_highest")
	suite.Nil(err)
}

func (suite *SkycoinSuit) TestConnectionsHighest() {
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
	found, err = util.FoundMetric(respBody, "connections_highest")
	suite.Nil(err)
	suite.True(found)
	var val util.NumericVec
	val, err = util.GetNumericVecValues(respBody, "connections_highest")
	suite.Nil(err)
	haveLabel := func(key string, values util.NumericVec) bool {
		for _, value := range values.Values {
			for _, label := range value.Labels {
				if label.Name == key {
					return true
				}
			}
		}
		return false
	}
	suite.True(haveLabel("Address", val))
}

func (suite *SkycoinSuit) TestConnectionsBurnFactor() {
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
	found, err = util.FoundMetric(respBody, "connections_burn_factor_hist")
	suite.Nil(err)
	suite.True(found)
	// var val float64
	_, err = util.GetHistogramValue(respBody, "connections_burn_factor_hist")
	suite.Nil(err)
}

func (suite *SkycoinSuit) TestConnectionsMaxTransactionSize() {
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
	found, err = util.FoundMetric(respBody, "connections_max_transaction_size")
	suite.Nil(err)
	suite.True(found)
	var val util.NumericVec
	val, err = util.GetNumericVecValues(respBody, "connections_max_transaction_size")
	suite.Nil(err)
	haveLabel := func(key string, values util.NumericVec) bool {
		for _, value := range values.Values {
			for _, label := range value.Labels {
				if label.Name == key {
					return true
				}
			}
		}
		return false
	}
	suite.True(haveLabel("Address", val))
}

func (suite *SkycoinSuit) TestConnectionsMaxDecimals() {
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
	found, err = util.FoundMetric(respBody, "connections_max_decimals")
	suite.Nil(err)
	suite.True(found)
	var val util.NumericVec
	val, err = util.GetNumericVecValues(respBody, "connections_max_decimals")
	suite.Nil(err)
	haveLabel := func(key string, values util.NumericVec) bool {
		for _, value := range values.Values {
			for _, label := range value.Labels {
				if label.Name == key {
					return true
				}
			}
		}
		return false
	}
	suite.True(haveLabel("Address", val))
}
