package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// RootConfig implements config.RextRoot
type RootConfig struct {
	services []config.RextServiceDef
}

// Clone make a deep copy of RootConfig or return an error if any
func (root RootConfig) Clone() (cRc config.RextRoot, err error) {
	var cSrvs []config.RextServiceDef
	for _, srv := range root.GetServices() {
		var cSrv config.RextServiceDef
		if cSrv, err = srv.Clone(); err != nil {
			log.WithError(err).Errorln("can not clone services in root config")
			return cRc, err
		}
		cSrvs = append(cSrvs, cSrv)
	}
	cRc = NewRootConfig(cSrvs)
	return cRc, err
}

// GetServices return the services
func (root RootConfig) GetServices() []config.RextServiceDef {
	services := make([]config.RextServiceDef, len(root.services))
	for idxSrv := range root.services {
		services[idxSrv] = root.services[idxSrv]
	}
	return services
}

// AddService add a service
func (root *RootConfig) AddService(srv config.RextServiceDef) {
	root.services = append(root.services, srv)
}

// Validate the root, return true if any error is found
func (root RootConfig) Validate() bool {
	return config.ValidateRoot(&root)
}

// NewRootConfig return a new root config instance
func NewRootConfig(services []config.RextServiceDef) config.RextRoot {
	return &RootConfig{services: services}
}
