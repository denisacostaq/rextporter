package toml2config

import (
	"fmt"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/memconfig"
	"github.com/simelo/rextporter/src/tomlconfig"
	log "github.com/sirupsen/logrus"
)

type metricName2Metric map[string]tomlconfig.Metric
type serviceName2MetricName2Metric map[string]metricName2Metric

func buildMetricsMapping(conf tomlconfig.RootConfig) (metricsMapping serviceName2MetricName2Metric) {
	metricsMapping = make(serviceName2MetricName2Metric)
	for _, srv := range conf.Services {
		mtrName2Metric := make(metricName2Metric)
		for _, mtr := range srv.Metrics {
			mtrName2Metric[mtr.Name] = mtr
		}
		metricsMapping[srv.Name] = mtrName2Metric
	}
	return metricsMapping
}

func createService(srv tomlconfig.Service, metricsMapping serviceName2MetricName2Metric) (service config.RextServiceDef, err error) {
	mtrN2Metric := metricsMapping[srv.Name]
	service = &memconfig.Service{}
	service.SetProtocol(srv.Protocol)
	basePath := fmt.Sprintf("%s://%s:%d", service.GetProtocol(), srv.Location.Location, srv.Port)
	service.SetBasePath(basePath)
	srvOpts := service.GetOptions()
	if _, err = srvOpts.SetString(config.OptKeyRextServiceDefJobName, srv.Name); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextServiceDefJobName, "val": srv.Name}).Errorln("error saving job name")
		return service, err
	}
	if _, err = srvOpts.SetString(config.OptKeyRextServiceDefInstanceName, fmt.Sprintf("%s:%d", srv.Location.Location, srv.Port)); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextServiceDefInstanceName, "val": fmt.Sprintf("%s:%d", srv.Location.Location, srv.Port)}).Errorln("error saving instance name")
		return service, err
	}
	auth := &memconfig.HTTPAuth{}
	auth.SetAuthType(srv.AuthType)
	authOpts := auth.GetOptions()
	if _, err = authOpts.SetString(config.OptKeyRextAuthDefTokenHeaderKey, srv.TokenHeaderKey); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextAuthDefTokenHeaderKey, "val": srv.TokenHeaderKey}).Errorln("error saving token header key")
		return service, err
	}
	if _, err = authOpts.SetString(config.OptKeyRextAuthDefTokenKeyFromEndpoint, srv.TokenKeyFromEndpoint); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextAuthDefTokenKeyFromEndpoint, "val": srv.TokenKeyFromEndpoint}).Errorln("error saving token key from endpoint")
		return service, err
	}
	if _, err = authOpts.SetString(config.OptKeyRextAuthDefTokenGenEndpoint, srv.GenTokenEndpoint); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextAuthDefTokenGenEndpoint, "val": srv.GenTokenEndpoint}).Errorln("error saving token endpoint")
		return service, err
	}
	service.SetAuthForBaseURL(auth)
	for _, resPath := range srv.ResourcePaths {
		var resDef config.RextResourceDef
		switch resPath.PathType {
		case "rest_api":
			resDef = createResourceFrom4API(mtrN2Metric, resPath)
			resDef.SetType(resPath.PathType)
			resDef.SetResourceURI(resPath.Path)
			decoder := memconfig.NewDecoder(resPath.PathType, nil)
			resDef.SetDecoder(decoder)
			resOpts := resDef.GetOptions()
			// FIXME(denisacostaq@gmail.com): OptKeyRextResourceDefHTTPMethod should be inside the service or the resource
			if _, err = resOpts.SetString(config.OptKeyRextResourceDefHTTPMethod, resPath.HTTPMethod); err != nil {
				log.WithFields(log.Fields{"key": config.OptKeyRextResourceDefHTTPMethod, "val": resPath.HTTPMethod}).Errorln("error saving http method")
				return service, err
			}
		case "metrics_fordwader":
			resDef = createResourceFrom4ExposedMetrics(resPath)
			decoder := memconfig.NewDecoder(resPath.PathType, nil)
			resDef.SetDecoder(decoder)
		default:
			log.WithField("resource_path_type", resPath.PathType).Errorln("valid types are rest_api or metrics_fordwader")
			return service, config.ErrKeyInvalidType
		}
		service.AddResource(resDef)
	}
	return service, err
}

func createResourceFrom4API(mtrN2Metric map[string]tomlconfig.Metric, resPath tomlconfig.ResourcePath) (resDef config.RextResourceDef) {
	resDef = &memconfig.ResourceDef{}
	resDef.SetType(resPath.PathType)
	resDef.SetResourceURI(resPath.Path)
	resOpts := resDef.GetOptions()
	if _, err := resOpts.SetString(config.OptKeyRextResourceDefHTTPMethod, resPath.HTTPMethod); err != nil {
		log.WithFields(log.Fields{"key": config.OptKeyRextResourceDefHTTPMethod, "val": resPath.HTTPMethod}).Errorln("error saving http method")
		return resDef
	}
	for _, mtrName := range resPath.MetricNames {
		mtr /*, foundMetric*/ := mtrN2Metric[mtrName]
		// if !foundMetric {
		// 	continue
		// }
		nodeSolver := &memconfig.NodeSolver{MType: resPath.NodeSolverType}
		nodeSolver.SetNodePath(mtr.Path)
		metric := &memconfig.MetricDef{}
		metric.SetMetricName(mtr.Name)
		metric.SetMetricType(mtr.Options.Type)
		metric.SetMetricDescription(mtr.Options.Description)
		mtrOpts := metric.GetOptions()
		for _, tomlLabel := range mtr.Options.Labels {
			label := &memconfig.LabelDef{}
			label.SetName(tomlLabel.Name)
			lns := &memconfig.NodeSolver{MType: "jsonNode"}
			lns.SetNodePath(tomlLabel.Path)
			label.SetNodeSolver(lns)
			metric.AddLabel(label)
		}
		if mtr.Options.Type == config.KeyMetricTypeHistogram {
			if _, err := mtrOpts.SetObject(config.OptKeyRextMetricDefHMetricBuckets, mtr.HistogramOptions.Buckets); err != nil {
				log.WithFields(log.Fields{"key": config.OptKeyRextMetricDefHMetricBuckets, "value": mtr.HistogramOptions.Buckets}).Errorln("error saving buckets for histogram")
				return resDef
			}
		}
		metric.SetNodeSolver(nodeSolver)
		resDef.AddMetricDef(metric)
	}
	return resDef
}

func createResourceFrom4ExposedMetrics(resPath tomlconfig.ResourcePath) (resDef config.RextResourceDef) {
	resDef = &memconfig.ResourceDef{}
	resDef.SetType(resPath.PathType)
	resDef.SetResourceURI(resPath.Path)
	return resDef
}

// Fill receive a tomlconfig.RootConfig and return an equivalent config.RextRoot
func Fill(conf tomlconfig.RootConfig) (root config.RextRoot, err error) {
	root = &memconfig.RootConfig{}
	metricsMapping := buildMetricsMapping(conf)
	for _, srv := range conf.Services {
		var service config.RextServiceDef
		if service, err = createService(srv, metricsMapping); err != nil {
			log.WithError(err).Errorln("can not fill service info")
			return root, err
		}
		root.AddService(service)
	}
	if root.Validate() {
		err = config.ErrKeyConfigHaveSomeErrors
		return root, err
	}
	return root, err
}
