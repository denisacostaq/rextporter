package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/template"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/toml2config"
	"github.com/simelo/rextporter/src/tomlconfig"
	"github.com/simelo/rextporter/test/util/testrand"
	log "github.com/sirupsen/logrus"
)

const mainConfigFileContenTemplate = `configTransport = "file" # "file" | "consulCatalog"
# render a template with a portable path
servicesConfigPath = "{{.ServicesConfPath}}"
metricsForServicesConfigPath = "{{.MetricsForServicesConfPath}}"
resourcePathsForServicesConfPath = "{{.ResourcePathsForServicesConfPath}}"
`

const servicesConfigFileContenTemplate = `# Service configuration.{{range .Services}}
[[services]]
	name = "{{.Name}}"
	protocol = "http"
	port = {{.Port}}
	basePath = "{{.BasePath}}"
	authType = "CSRF"
	tokenHeaderKey = "X-CSRF-Token"
	genTokenEndpoint = "/api/v1/csrf"
	tokenKeyFromEndpoint = "csrf_token"
		
	[services.location]
		location = "localhost"

{{end}}
`

const metricsConfigFileContenTemplate = `# All metrics to be measured.{{range .Metrics}}
[[metrics]]
	name = "{{.Name}}"
	path = "{{.Path}}"
	nodeSolver = "ns0132" # FIXME(denisacostaq@gmail.com): make portable

	[metrics.options]
		type = "{{.Options.Type}}"
		description = "{{.Options.Description}}"{{if gt (len .Options.Labels) 0}} {{$len := (len .Options.Labels)}}
		[[metrics.options.labels]]{{range .Options.Labels}}
			name = "{{.Name}}"
			path = "{{.Path}}"
{{end}}{{end}}
	
	[metrics.histogramOptions]{{if gt (len .HistogramOptions.Buckets) 0}} {{$len := (len .HistogramOptions.Buckets)}}
		buckets = [{{range $i, $v := .HistogramOptions.Buckets}}{{$v}}{{if lt (inc $i) $len}}, {{end}}{{end}}]{{end}}{{if gt (len .HistogramOptions.ExponentialBuckets) 0}} {{$len := (len .HistogramOptions.ExponentialBuckets)}}
		exponentialBuckets = [{{range $i, $v := .HistogramOptions.ExponentialBuckets}}{{$v}}{{if lt (inc $i) $len}}, {{end}}{{end}}]
{{end}}
{{end}}
`

const metricsForServicesConfFileContenTemplate = `metricPathsForServicesConfig = [{{range $key, $value := .}}
	{ {{$key}} = "{{$value}}" },{{end}}
]
`

const resourceForServicesConfFileContentTemplate = `resourcePathsForServicesConfig = [{{range $key, $value := .}}
	{ {{$key}} = "{{$value}}" },{{end}}
]
`

const serviceResourcePathsFileContentTemplate = `{{range .}}[[ResourcePaths]]
	Name = "{{.Name}}"
	Path = "{{.Path}}"
	PathType = "{{.PathType}}"{{if ne .NodeSolverType ""}}
	nodeSolverType = "{{.NodeSolverType}}"{{end}}{{if gt (len .MetricNames) 0}} {{$len := (len .MetricNames)}}
	MetricNames = [{{range $i, $v := .MetricNames}}"{{$v}}"{{if lt (inc $i) $len}}, {{end}}{{end}}]
{{end}}	

{{end}}
`

func createConfigFile(tmplContent, path string, data interface{}) (err error) {
	if len(tmplContent) == 0 || len(path) == 0 {
		log.Errorln("template content should not be empty")
		return config.ErrKeyEmptyValue
	}
	tmpl := template.New("fileConfig")
	var templateEngine *template.Template
	funcs := template.FuncMap{"inc": func(i int) int { return i + 1 }}
	if templateEngine, err = tmpl.New("").Funcs(funcs).Parse(tmplContent); err != nil {
		log.WithField("template", tmplContent).Errorln("Can not parse template content")
		return config.ErrKeyDecodingFile
	}
	var configFile *os.File
	if configFile, err = os.Create(path); err != nil {
		log.WithFields(log.Fields{"err": err, "path": path}).Errorln("error creating config file")
		return ErrKeyWritingFsStructure
	}
	if err = templateEngine.Execute(configFile, data); err != nil {
		log.WithFields(log.Fields{"err": err, "data": data}).Errorln("error writing config file")
		return ErrKeyWritingFile
	}
	return nil
}

var (
	// ErrKeyWritingFsStructure tells about a fs change(creating a file and/or folder)
	ErrKeyWritingFsStructure = errors.New("Error creating file/folder")
	// ErrKeyWritingFile tells about a write error in a file
	ErrKeyWritingFile = errors.New("Error writing file")
)

func createFullConfig(mainConfFilePath string, conf tomlconfig.RootConfig) (err error) {
	srvsConfDir := testrand.RFolderPath()
	mtrs4ServiceConfDir := testrand.RFolderPath()
	res4ServiceConfDir := testrand.RFolderPath()
	dirs := []string{srvsConfDir, mtrs4ServiceConfDir, res4ServiceConfDir}
	if err = createDirectoriesWithFullDepth(dirs); err != nil {
		log.WithError(err).Errorln("error creating directory")
		return ErrKeyWritingFsStructure
	}
	srvsConfPath := filepath.Join(srvsConfDir, testrand.RName())
	mtrs4ServiceConfPath := filepath.Join(mtrs4ServiceConfDir, testrand.RName())
	res4ServiceConfPath := filepath.Join(res4ServiceConfDir, testrand.RName())
	type mainConfigData struct {
		ServicesConfPath                 string
		MetricsForServicesConfPath       string
		ResourcePathsForServicesConfPath string
	}
	confData := mainConfigData{
		ServicesConfPath:                 srvsConfPath,
		MetricsForServicesConfPath:       mtrs4ServiceConfPath,
		ResourcePathsForServicesConfPath: res4ServiceConfPath,
	}
	if e := createConfigFile(mainConfigFileContenTemplate, mainConfFilePath, confData); e != nil {
		log.WithError(e).Errorln("error writing main config")
		err = ErrKeyWritingFsStructure
	}
	if e := createConfigFile(servicesConfigFileContenTemplate, srvsConfPath, conf); e != nil {
		log.WithError(e).Errorln("error writing service config")
		err = ErrKeyWritingFsStructure
	}
	mtrs4Srvs := make(map[string]string)
	res4Srvs := make(map[string]string)
	for _, srv := range conf.Services {
		mtrsConfDir := testrand.RFolderPath()
		res4SrvsConfDir := testrand.RFolderPath()
		dirs = []string{mtrsConfDir, res4SrvsConfDir}
		if e := createDirectoriesWithFullDepth(dirs); e != nil {
			log.WithError(e).Errorln("error creating directory")
			return ErrKeyWritingFsStructure
		}
		mtrsConfPath := filepath.Join(mtrsConfDir, testrand.RName())
		res4SrvsConfPath := filepath.Join(res4SrvsConfDir, testrand.RName())
		mtrs4Srvs[srv.Name] = mtrsConfPath
		res4Srvs[srv.Name] = res4SrvsConfPath
		if e := createConfigFile(metricsConfigFileContenTemplate, mtrsConfPath, srv); e != nil {
			log.WithError(e).Errorln("error writing metrics config")
			err = ErrKeyWritingFsStructure
		}
		if e := createConfigFile(serviceResourcePathsFileContentTemplate, res4SrvsConfPath, srv.ResourcePaths); e != nil {
			log.WithError(e).Errorln("error writing service resource paths config")
			err = ErrKeyWritingFsStructure
		}
	}
	if e := createConfigFile(resourceForServicesConfFileContentTemplate, res4ServiceConfPath, res4Srvs); e != nil {
		log.WithError(e).Errorln("error writing resources paths for services config")
		err = ErrKeyWritingFsStructure
	}
	if e := createConfigFile(metricsForServicesConfFileContenTemplate, mtrs4ServiceConfPath, mtrs4Srvs); e != nil {
		log.WithError(e).Errorln("error writing metrics for service config")
		err = ErrKeyWritingFsStructure
	}
	return err
}

func createDirectoriesWithFullDepth(dirs []string) (err error) {
	for _, dir := range dirs {
		if err = os.MkdirAll(dir, 0750); err != nil {
			log.WithFields(log.Fields{"err": err, "dir": dir}).Errorln("Error creating directory")
			err = ErrKeyWritingFsStructure
		}
	}
	return err
}

func getConfig(mainConfFilePath string) (rootConf config.RextRoot, err error) {
	rawConf, err := tomlconfig.ReadConfigFromFileSystem(mainConfFilePath)
	if err != nil {
		log.WithField("path", mainConfFilePath).Errorln("error reading config from file system")
		return rootConf, err
	}
	if rootConf, err = toml2config.Fill(rawConf); err != nil {
		log.WithField("conf", rawConf).Errorln("error filling config info")
		return rootConf, err
	}
	return rootConf, err
}

func readListenPortFromFile() (port uint16, err error) {
	var path string
	path, err = testrand.FilePathToSharePort()
	var file *os.File
	file, err = os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		log.WithError(err).Errorln("error opening file")
		return 0, err
	}
	defer file.Close()
	_, err = fmt.Fscanf(file, "%d", &port)
	if err != nil {
		log.WithError(err).Errorln("error reading file")
		return port, err
	}
	return port, err
}
