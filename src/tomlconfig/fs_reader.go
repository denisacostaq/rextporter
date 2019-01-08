package tomlconfig

import (
	"errors"

	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// ErrKeyReadingFile check for empty and/or null
	ErrKeyReadingFile = errors.New("Error reading file")
)

type configFromFile struct {
	filePath string
}

type mainConfig struct {
	ServicesConfigTransport          string
	ServicesConfigPath               string
	MetricsForServicesConfigPath     string
	ResourcePathsForServicesConfPath string
}

func (cf configFromFile) readTomlFile(data interface{}) error {
	if len(cf.filePath) == 0 {
		log.Errorln("file path is required to read toml config")
		return config.ErrKeyEmptyValue
	}
	viper.SetConfigType("toml")
	viper.SetConfigFile(cf.filePath)
	if err := viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{"err": err, "path": cf.filePath}).Errorln("error reading toml config file")
		return ErrKeyReadingFile
	}
	if err := viper.Unmarshal(data); err != nil {
		log.WithFields(log.Fields{"err": err, "path": cf.filePath}).Errorln("Error decoding toml config file content")
		return ErrKeyReadingFile
	}
	return nil
}

func (cf configFromFile) readMainConf() (mainConf mainConfig, err error) {
	if err = cf.readTomlFile(&mainConf); err != nil {
		log.Errorln("error reading main config")
		return mainConf, err
	}
	return mainConf, err
}

func (cf configFromFile) readServicesConf() (services []Service, err error) {
	var root RootConfig
	if err = cf.readTomlFile(&root); err != nil {
		log.Errorln("error reading services config")
		return services, err
	}
	services = root.Services
	return services, err
}

func (cf configFromFile) readMetricsForServiceConf() (metricsConf MetricsTemplate, err error) {
	type metricsForServiceConfig struct {
		Metrics MetricsTemplate
	}
	var metricsForServiceConf metricsForServiceConfig
	if err = cf.readTomlFile(&metricsForServiceConf); err != nil {
		log.Errorln("error reading metrics config")
		return metricsConf, err
	}
	metricsConf = metricsForServiceConf.Metrics
	return metricsConf, err
}

func (cf configFromFile) readResourcePathsForServiceConf() (resPaths4Service ResourcePathTemplate, err error) {
	type resourcePathsForServiceConfig struct {
		ResourcePaths ResourcePathTemplate
	}
	var resourcePathsForServiceConf resourcePathsForServiceConfig
	if err = cf.readTomlFile(&resourcePathsForServiceConf); err != nil {
		log.Errorln("error reading resource path for services config")
		return resPaths4Service, err
	}
	resPaths4Service = resourcePathsForServiceConf.ResourcePaths
	return resPaths4Service, err
}

func (cf configFromFile) readResourcePathsForServicesConf() (resPaths4Services map[string]string, err error) {
	type resourcePathsForServicesConfig struct {
		ResourcePathsForServicesConfig map[string]string
	}
	var resourcePathsForServicesConf resourcePathsForServicesConfig
	if err = cf.readTomlFile(&resourcePathsForServicesConf); err != nil {
		log.Errorln("error reading main config")
		return resPaths4Services, err
	}
	resPaths4Services = resourcePathsForServicesConf.ResourcePathsForServicesConfig
	return resPaths4Services, err
}

func (cf configFromFile) readMetricsPathsForServicesConf() (mtrPaths4Services map[string]string, err error) {
	type metricPathsForServicesConfig struct {
		MetricPathsForServicesConfig map[string]string
	}
	var metricPathsForServicesConf metricPathsForServicesConfig
	if err = cf.readTomlFile(&metricPathsForServicesConf); err != nil {
		log.Errorln("error reading metric for services config")
		return mtrPaths4Services, err
	}
	mtrPaths4Services = metricPathsForServicesConf.MetricPathsForServicesConfig
	return mtrPaths4Services, err
}

func readRootStructure(mainConf mainConfig) (rootConf RootConfig, err error) {
	srvConfReader := configFromFile{filePath: mainConf.ServicesConfigPath}
	rootConf.Services, err = srvConfReader.readServicesConf()
	if err != nil {
		log.WithError(err).Errorln("error reading services config")
		return rootConf, ErrKeyReadingFile
	}
	resPaths4ServicesReader := configFromFile{filePath: mainConf.ResourcePathsForServicesConfPath}
	var resPath4Service, mtrPath4Service map[string]string
	if resPath4Service, err = resPaths4ServicesReader.readResourcePathsForServicesConf(); err != nil {
		log.WithError(err).Errorln("error reading resource paths for services config")
		return rootConf, ErrKeyReadingFile
	}
	mtrPaths4ServicesReader := configFromFile{filePath: mainConf.MetricsForServicesConfigPath}
	if mtrPath4Service, err = mtrPaths4ServicesReader.readMetricsPathsForServicesConf(); err != nil {
		log.WithError(err).Errorln("error reading metric paths for services config")
		return rootConf, ErrKeyReadingFile
	}
	for idxService, service := range rootConf.Services {
		resPath4ServiceReader := configFromFile{filePath: resPath4Service[service.Name]}
		if rootConf.Services[idxService].ResourcePaths, err = resPath4ServiceReader.readResourcePathsForServiceConf(); err != nil {
			log.WithFields(log.Fields{"err": err, "service": service.Name}).Warnln("error reading resource paths for service")
			return rootConf, ErrKeyReadingFile
		}
		mtrPath4ServiceReader := configFromFile{filePath: mtrPath4Service[service.Name]}
		if rootConf.Services[idxService].Metrics, err = mtrPath4ServiceReader.readMetricsForServiceConf(); err != nil {
			log.WithFields(log.Fields{"err": err, "service": service.Name}).Warnln("error reading resource paths for service")
			return rootConf, ErrKeyReadingFile
		}
	}
	return rootConf, err
}

// ReadConfigFromFileSystem will read the config from the file system.
func ReadConfigFromFileSystem(filePath string) (rootConf RootConfig, err error) {
	// TODO(denisacostaq@gmail.com): use configlocator.MainFile() here if filePath is empty
	mainConfigPath := filePath
	mainConfReader := configFromFile{filePath: mainConfigPath}
	var mainConf mainConfig
	if mainConf, err = mainConfReader.readMainConf(); err != nil {
		log.WithError(err).Errorln("error reading main config file")
		return rootConf, ErrKeyReadingFile
	}
	if rootConf, err = readRootStructure(mainConf); err != nil {
		log.WithError(err).Errorln("error reading root structure conf")
		return rootConf, err
	}
	return rootConf, nil
}
