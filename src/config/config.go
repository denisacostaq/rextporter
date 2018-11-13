package config

import (
	"bytes"
	"container/list"
	"fmt"
	"net/url"
	"strings"

	"github.com/simelo/rextporter/src/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// RootConfig is the top level node for the config tree, it has a list of metrics and a
// service from which get this metrics.
type RootConfig struct {
	Services []Service `json:"services"`
}

var rootConfig RootConfig

// Config TODO(denisacostaq@gmail.com): make a singleton
func Config() RootConfig {
	//if b, err := json.MarshalIndent(rootConfig, "", " "); err != nil {
	//	log.Println("Error marshaling:", err)
	//} else {
	//	os.Stdout.Write(b)
	//	log.Println("\n\n\n\n\n")
	//}
	// TODO(denisacostaq@gmail.com): Make it a singleton
	return rootConfig
}

// NewConfigFromRawString allow you to define a `.toml` config in the fly, a raw string with the "config content"
func NewConfigFromRawString(strConf string) error {
	const generalScopeErr = "error creating a config instance"
	viper.SetConfigType("toml")
	buff := bytes.NewBuffer([]byte(strConf))
	if err := viper.ReadConfig(buff); err != nil {
		errCause := fmt.Sprintln("can not read the buffer: ", err.Error())
		return common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	rootConfig = RootConfig{}
	if err := viper.Unmarshal(&rootConfig); err != nil {
		errCause := fmt.Sprintln("can not decode the config data: ", err.Error())
		return common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	rootConfig.validate()
	return nil
}

// newMetricsConfig desserialize a metrics config from the 'toml' file path
func newMetricsConfig(path string) (metricsConf []Metric, err error) {
	const generalScopeErr = "error reading metrics config"
	if strings.Compare(path, "") == 0 {
		errCause := "path should not be null"
		return metricsConf, common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		errCause := fmt.Sprintln("error reading config file: ", path, err.Error())
		return metricsConf, common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	type metricsForService struct {
		Metrics []Metric
	}
	var root metricsForService
	if err := viper.Unmarshal(&root); err != nil {
		errCause := fmt.Sprintln("can not decode the config data: ", err.Error())
		return metricsConf, common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	metricsConf = root.Metrics
	return metricsConf, nil
}

// newServiceConfigFromFile desserialize a service config from the 'toml' file path
func newServiceConfigFromFile(path string, conf mainConfigData) (servicesConf []Service, err error) {
	const generalScopeErr = "error reading service config"
	serviceConfReader := NewServiceConfigFromFile(path)
	if servicesConf, err = serviceConfReader.GetConfig(); err != nil {
		errCause := "error reading service config"
		return servicesConf, common.ErrorFromThisScope(errCause, generalScopeErr)
	}
	for idxService, service := range servicesConf {
		if servicesConf[idxService].Metrics, err = newMetricsConfig(conf.MetricsConfigPath(service.Name)); err != nil {
			errCause := "error reading metrics config: " + err.Error()
			panic(common.ErrorFromThisScope(errCause, generalScopeErr))
		}
	}
	return servicesConf, err
}

// NewConfigFromFileSystem will read the config from the file system, you should send the
// metric config file path and service config file path into metricsPath, servicePath respectively.
// This function can cause a panic.
// TODO(denisacostaq@gmail.com): make this a singleton
func NewConfigFromFileSystem(mainConfigPath string) {
	const generalScopeErr = "error getting config values from file system"
	var conf mainConfigData
	var err error
	if conf, err = newMainConfigData(mainConfigPath); err != nil {
		errCause := "error reading metrics config: " + err.Error()
		panic(errCause)
	}
	if rootConfig.Services, err = newServiceConfigFromFile(conf.ServiceConfigPath(), conf); err != nil {
		errCause := "root cause: " + err.Error()
		panic(common.ErrorFromThisScope(errCause, generalScopeErr))
	}
	rootConfig.validate()
}

// FilterMetricsByType will return all the metrics who match whit the 't' parameter.
func (conf RootConfig) FilterMetricsByType(t string) (metrics []Metric) {
	tmpMetrics := list.New()
	for _, service := range conf.Services {
		for _, metric := range service.Metrics {
			if strings.Compare(metric.Options.Type, t) == 0 {
				tmpMetrics.PushBack(metric)
			}
		}
	}
	metrics = make([]Metric, tmpMetrics.Len())
	idxLink := 0
	for it := tmpMetrics.Front(); it != nil; it = it.Next() {
		metrics[idxLink] = it.Value.(Metric)
		idxLink++
	}
	return metrics
}

// FilterServicesByType will return all the services who match whit the 't' parameter.
func (conf RootConfig) FilterServicesByType(t string) (services []Service) {
	tmpServices := list.New()
	for _, service := range conf.Services {
		if strings.Compare(service.Mode, t) == 0 {
			tmpServices.PushBack(service)
		}
	}
	services = make([]Service, tmpServices.Len())
	idxLink := 0
	for it := tmpServices.Front(); it != nil; it = it.Next() {
		services[idxLink] = it.Value.(Service)
		idxLink++
	}
	return services
}

func (conf RootConfig) validate() {
	var errs []error
	for _, service := range conf.Services {
		errs = append(errs, service.validate()...)
	}
	if len(errs) != 0 {
		defer log.Panicln("some errors found")
		for _, err := range errs {
			log.WithError(err).Errorln("Error")
		}
	}
}

// isValidUrl tests a string to determine if it is a valid URL or not.
func isValidURL(toTest string) bool {
	if _, err := url.ParseRequestURI(toTest); err != nil {
		return false
	}
	return true
}
