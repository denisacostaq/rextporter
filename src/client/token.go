package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// TokenCreator create token clients
type TokenCreator baseFactory

// CreateClient create a token client
func (tc TokenCreator) CreateClient() (cl Client, err error) {
	const generalScopeErr = "error creating a client to get a toke from remote endpoint for making future requests"
	// client = TokenClient{baseClient: baseClient{service: service}}
	// FIXME(denisacostaq@gmail.com): make the "GET" configurable.
	var req *http.Request
	if req, err = http.NewRequest("GET", tc.dataSource, nil); err != nil {
		errCause := fmt.Sprintln("can not create the request: ", err.Error())
		return cl, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	cl = TokenClient{
		baseClient: baseClient{ // nolint megacheck
			jobName:                        tc.jobName,
			instanceName:                   tc.instanceName,
			dataSource:                     tc.dataSource,
			dataSourceResponseDurationDesc: tc.dataSourceResponseDurationDesc,
		},
		req: req,
	}
	return cl, nil
}

// TokenClient implements the getRemoteInfo method from `client.Client` interface by using some .toml config parameters
// like for example: where is the host? it should be a GET, a POST or some other ... It works like an http wrapper to
// to get a token from the server.
// sa newTokenClient method.
type TokenClient struct {
	baseClient
	req *http.Request
}

// GetData can get a token value from a remote server
func (client TokenClient) GetData(metricsCollector chan<- prometheus.Metric) (data []byte, err error) {
	const generalScopeErr = "error making a server request to get token from remote endpoint"
	httpClient := &http.Client{}
	var resp *http.Response
	{
		successResponse := false
		defer func(startTime time.Time) {
			duration := time.Since(startTime).Seconds()
			labels := []string{client.jobName, client.instanceName, client.dataSource}
			if successResponse {
				if _, err := prometheus.NewConstMetric(client.dataSourceResponseDurationDesc, prometheus.GaugeValue, duration, labels...); err == nil {
					// FIXME(denisacostaq@gmail.com): this approach may not work for auth because multiple calls
					// metricsCollector <- metric
				} else {
					log.WithFields(log.Fields{"err": err, "labels": labels}).Errorln("can not send dataSource response duration resolving token")
					return
				}
			}
		}(time.Now().UTC())
		if resp, err = httpClient.Do(client.req); err != nil {
			errCause := fmt.Sprintln("can not do the request: ", err.Error())
			return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		if resp.StatusCode != http.StatusOK {
			errCause := fmt.Sprintf("no success response, status %s", resp.Status)
			return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		successResponse = true
	}
	defer resp.Body.Close()
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		errCause := fmt.Sprintln("can not read the body: ", err.Error())
		return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	return data, nil
}
