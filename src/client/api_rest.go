package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/oliveagle/jsonpath"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// APIRestCreator have info to create api rest an client
type APIRestCreator struct {
	baseFactory
	httpMethod           string
	dataPath             string
	tokenPath            string
	tokenHeaderKey       string
	tokenKeyFromEndpoint string
}

// CreateAPIRestCreator create an APIRestCreator
func CreateAPIRestCreator(resConf config.RextResourceDef, srvConf config.RextServiceDef, dataSourceResponseDurationDesc *prometheus.Desc) (cf CacheableFactory, err error) {
	resOptions := resConf.GetOptions()
	httpMethod, err := resOptions.GetString(config.OptKeyRextResourceDefHTTPMethod)
	if err != nil {
		log.WithError(err).Errorln("Can not find httpMethod")
		return cf, err
	}
	resURI := strings.TrimPrefix(resConf.GetResourcePATH(srvConf.GetBasePath()), srvConf.GetBasePath())
	auth := resConf.GetAuth(srvConf.GetAuthForBaseURL())
	var tkHeaderKey, tkKeyFromEndpoint, tkKeyGenEndpoint string
	if auth != nil {
		authOpts := auth.GetOptions()
		tkHeaderKey, err = authOpts.GetString(config.OptKeyRextAuthDefTokenHeaderKey)
		if err != nil {
			log.WithError(err).Errorln("Can not find tokenHeaderKey")
			return cf, err
		}
		tkKeyFromEndpoint, err = authOpts.GetString(config.OptKeyRextAuthDefTokenKeyFromEndpoint)
		if err != nil {
			log.WithError(err).Errorln("Can not find tokenKeyFromEndpoint")
			return cf, err
		}
		tkKeyGenEndpoint, err = authOpts.GetString(config.OptKeyRextAuthDefTokenGenEndpoint)
		if err != nil {
			log.WithError(err).Errorln("Can not find tkKeyGenEndpoint")
			return cf, err
		}
	} else {
		log.Warnln("you have an empty auth")
	}
	srvOpts := srvConf.GetOptions()
	jobName, err := srvOpts.GetString(config.OptKeyRextServiceDefJobName)
	if err != nil {
		log.WithError(err).Errorln("Can not find jobName")
		return cf, err
	}
	instanceName, err := srvOpts.GetString(config.OptKeyRextServiceDefInstanceName)
	if err != nil {
		log.WithError(err).Errorln("Can not find instanceName")
		return cf, err
	}
	cf = APIRestCreator{
		baseFactory: baseFactory{
			jobName:                        jobName,
			instanceName:                   instanceName,
			dataSource:                     resURI,
			dataSourceResponseDurationDesc: dataSourceResponseDurationDesc,
		},
		httpMethod:           httpMethod,
		dataPath:             resConf.GetResourcePATH(srvConf.GetBasePath()),
		tokenPath:            srvConf.GetBasePath() + tkKeyGenEndpoint,
		tokenHeaderKey:       tkHeaderKey,
		tokenKeyFromEndpoint: tkKeyFromEndpoint,
	}
	return cf, err
}

// CreateClient create an api rest client
func (ac APIRestCreator) CreateClient() (cl CacheableClient, err error) {
	const generalScopeErr = "error creating api rest client"
	var req *http.Request
	if req, err = http.NewRequest(ac.httpMethod, ac.dataPath, nil); err != nil {
		errCause := fmt.Sprintln("can not create the request client: ", err.Error())
		return APIRest{}, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var tokenClient Client
	tc := TokenCreator{
		jobName:                        ac.jobName,
		instanceName:                   ac.instanceName,
		dataSource:                     ac.tokenPath,
		dataSourceResponseDurationDesc: ac.dataSourceResponseDurationDesc,
	}
	if tokenClient, err = tc.CreateClient(); err != nil {
		errCause := fmt.Sprintln("create token client: ", err.Error())
		return APIRest{}, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	cl = APIRest{
		baseClient: baseClient{
			jobName:                        ac.jobName,
			instanceName:                   ac.instanceName,
			dataSource:                     ac.dataSource,
			dataSourceResponseDurationDesc: ac.dataSourceResponseDurationDesc,
		},
		baseCacheableClient:  baseCacheableClient(ac.dataPath),
		req:                  req,
		tokenClient:          tokenClient,
		tokenHeaderKey:       ac.tokenHeaderKey,
		tokenKeyFromEndpoint: ac.tokenKeyFromEndpoint,
	}
	return cl, nil
}

// APIRest have common data to be shared through embedded struct in those type who implement the client.Client interface
type APIRest struct {
	baseClient
	baseCacheableClient
	req                  *http.Request
	tokenClient          Client
	tokenHeaderKey       string
	tokenKeyFromEndpoint string
	token                string
}

// GetData can retrieve data from a rest API with a retry pollicy for token expiration.
func (cl APIRest) GetData(metricsCollector chan<- prometheus.Metric) (data []byte, err error) {
	const generalScopeErr = "error making a server request to get metric from remote endpoint"
	cl.req.Header.Set(cl.tokenHeaderKey, cl.token)
	getData := func() (data []byte, err error) {
		httpClient := &http.Client{}
		var resp *http.Response
		{
			successResponse := false
			defer func(startTime time.Time) {
				duration := time.Since(startTime).Seconds()
				labels := []string{cl.jobName, cl.instanceName, cl.dataSource}
				if successResponse {
					if metric, err := prometheus.NewConstMetric(cl.dataSourceResponseDurationDesc, prometheus.GaugeValue, duration, labels...); err == nil {
						metricsCollector <- metric
					} else {
						log.WithFields(log.Fields{"err": err, "labels": labels}).Errorln("can not send dataSource response duration resolving api rest")
						return
					}
				}
			}(time.Now().UTC())
			if resp, err = httpClient.Do(cl.req); err != nil {
				log.WithFields(log.Fields{"err": err, "req": cl.req}).Errorln("no success response")
				errCause := fmt.Sprintln("can not do the request: ", err.Error())
				return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
			}
			if resp.StatusCode != http.StatusOK {
				log.WithFields(log.Fields{"status": resp.Status, "req": cl.req}).Errorln("no success response")
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
	if data, err = getData(); err != nil {
		// log.Println("can not do the request:", err.Error(), "trying with a new token...")
		if err = cl.resetToken(metricsCollector); err != nil {
			errCause := fmt.Sprintln("can not reset the token: ", err.Error())
			return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
		if data, err = getData(); err != nil {
			errCause := fmt.Sprintln("can not do the request after a token reset neither: ", err.Error())
			return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
		}
	}
	return data, nil
}

func (cl APIRest) resetToken(metricsCollector chan<- prometheus.Metric) (err error) {
	const generalScopeErr = "error making resetting the token"
	cl.token = ""
	var data []byte
	if data, err = cl.tokenClient.GetData(metricsCollector); err != nil {
		errCause := fmt.Sprintln("can make the request to get a token: ", err.Error())
		return util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var jsonData interface{}
	if err = json.Unmarshal(data, &jsonData); err != nil {
		errCause := fmt.Sprintln("can not decode the body: ", string(data), " ", err.Error())
		return util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	var val interface{}
	jPath := "$" + strings.Replace(cl.tokenKeyFromEndpoint, "/", ".", -1)
	if val, err = jsonpath.JsonPathLookup(jsonData, jPath); err != nil {
		errCause := fmt.Sprintln("can not locate the path: ", err.Error())
		return util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	tk, ok := val.(string)
	if !ok {
		errCause := fmt.Sprintln("unable the get the token as a string value")
		return util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	cl.token = tk
	if len(cl.token) == 0 {
		errCause := fmt.Sprintln("unable the get a not null(empty) token")
		return util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	return nil
}
