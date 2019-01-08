package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// Service implements config.RextServiceDef interface
type Service struct {
	basePath string
	// FIXME(denisacostaq@gmail.com): how to use base path, what about protocol, port, url
	protocol string
	auth     config.RextAuthDef
	// TODO(denisacostaq@gmail.com): rename to resources
	resources []config.RextResourceDef
	options   config.RextKeyValueStore
}

// Validate the service, return true if any error is found
func (srv Service) Validate() (hasError bool) {
	if srv.GetProtocol() == "http" {
		for _, res := range srv.GetResources() {
			resPath := res.GetResourcePATH(srv.GetBasePath())
			if !util.IsValidURL(resPath) {
				hasError = true
			}
		}
	}
	if config.ValidateService(&srv) {
		hasError = true
	}
	return hasError
}

// Clone make a deep copy of Service or return an error if any
func (srv Service) Clone() (cSrv config.RextServiceDef, err error) {
	var cAuth config.RextAuthDef
	if srv.auth != nil {
		if cAuth, err = srv.auth.Clone(); err != nil {
			log.WithError(err).Errorln("can not clone auth in service")
			return cSrv, err
		}
	}
	var cOpts config.RextKeyValueStore
	if cOpts, err = srv.GetOptions().Clone(); err != nil {
		log.WithError(err).Errorln("can not clone options in service")
		return cSrv, err
	}
	var cResources []config.RextResourceDef
	for _, resource := range srv.GetResources() {
		var cResource config.RextResourceDef
		if cResource, err = resource.Clone(); err != nil {
			log.WithError(err).Errorln("can not clone resources in service")
			return cSrv, err
		}
		cResources = append(cResources, cResource)
	}
	cSrv = NewServiceConf(srv.basePath, srv.protocol, cAuth, cResources, cOpts)
	return cSrv, err
}

// SetBasePath set the base path for the service
func (srv *Service) SetBasePath(path string) {
	srv.basePath = path
}

// GetProtocol return the service protocol
func (srv Service) GetProtocol() string {
	return srv.protocol
}

// GetBasePath return the base path
func (srv Service) GetBasePath() string {
	return srv.basePath
}

// SetProtocol set the protocol for the service
func (srv *Service) SetProtocol(protocol string) {
	srv.protocol = protocol
}

// SetAuthForBaseURL set an auth for the service
func (srv *Service) SetAuthForBaseURL(auth config.RextAuthDef) {
	srv.auth = auth
}

// GetAuthForBaseURL return the base auth
func (srv Service) GetAuthForBaseURL() config.RextAuthDef {
	return srv.auth
}

// AddResource add a resource
func (srv *Service) AddResource(source config.RextResourceDef) {
	srv.resources = append(srv.resources, source)
}

// AddResources add multiple resources
func (srv *Service) AddResources(sources ...config.RextResourceDef) {
	srv.resources = append(srv.resources, sources...)
}

// GetResources return the resources inside this service
func (srv Service) GetResources() []config.RextResourceDef {
	return srv.resources
}

// GetOptions return key/value pairs for extra options
func (srv *Service) GetOptions() config.RextKeyValueStore {
	if srv.options == nil {
		srv.options = NewOptionsMap()
	}
	return srv.options
}

// NewServiceConf create a new service
func NewServiceConf(basePath, protocol string, auth config.RextAuthDef, resources []config.RextResourceDef, options config.RextKeyValueStore) config.RextServiceDef {
	return &Service{
		basePath:  basePath,
		protocol:  protocol,
		auth:      auth,
		resources: resources,
		options:   options,
	}
}
