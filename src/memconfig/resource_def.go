package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// ResourceDef implements the interface config.RextResourceDef
type ResourceDef struct {
	mType       string
	resourceURI string
	auth        config.RextAuthDef
	decoder     config.RextDecoderDef
	metrics     []config.RextMetricDef
	options     config.RextKeyValueStore
}

// Clone make a deep copy of ResourceDef or return an error if any
func (rd ResourceDef) Clone() (cRd config.RextResourceDef, err error) {
	var cAuth config.RextAuthDef
	if rd.GetAuth(nil) != nil {
		if cAuth, err = rd.GetAuth(nil).Clone(); err != nil {
			log.WithError(err).Errorln("can not clone http auth in resource")
			return cRd, err
		}
	}
	var cDecoder config.RextDecoderDef
	if rd.GetDecoder() != nil {
		if cDecoder, err = rd.GetDecoder().Clone(); err != nil {
			log.WithError(err).Errorln("can not clone decoder in resource")
			return cRd, err
		}
	}
	var cMetrics []config.RextMetricDef
	for _, metric := range rd.GetMetricDefs() {
		var cMetric config.RextMetricDef
		if cMetric, err = metric.Clone(); err != nil {
			log.WithError(err).Errorln("can nor clone metrics in resource")
			return cRd, err
		}
		cMetrics = append(cMetrics, cMetric)
	}
	var cOpts config.RextKeyValueStore
	if cOpts, err = rd.GetOptions().Clone(); err != nil {
		log.WithError(err).Errorln("can not clone options in metric")
		return cRd, err
	}
	cRd = NewResourceDef(rd.GetType(), rd.resourceURI, cAuth, cMetrics, cDecoder, cOpts)
	return cRd, err
}

// GetResourcePATH return the resource pat against the service base path
func (rd ResourceDef) GetResourcePATH(basePath string) string {
	return basePath + rd.resourceURI
}

// GetType return the path type
func (rd ResourceDef) GetType() string {
	return rd.mType
}

// SetType set the type
func (rd *ResourceDef) SetType(t string) {
	rd.mType = t
}

// SetResourceURI set the resource path inside the service
func (rd *ResourceDef) SetResourceURI(uri string) {
	rd.resourceURI = uri
}

// GetAuth return the defAuth if not have a specific one for this resource
func (rd ResourceDef) GetAuth(defAuth config.RextAuthDef) config.RextAuthDef {
	if rd.auth == nil {
		return defAuth
	}
	return rd.auth
}

// SetAuth set an specific auth for this resource
func (rd *ResourceDef) SetAuth(auth config.RextAuthDef) {
	rd.auth = auth
}

// SetDecoder set a decoder for the service
func (rd *ResourceDef) SetDecoder(decoder config.RextDecoderDef) {
	rd.decoder = decoder
}

// GetDecoder return thedecoder for this service
func (rd ResourceDef) GetDecoder() config.RextDecoderDef {
	return rd.decoder
}

// AddMetricDef add a metric definition inside the resource
func (rd *ResourceDef) AddMetricDef(mtrDef config.RextMetricDef) {
	rd.metrics = append(rd.metrics, mtrDef)
}

// GetMetricDefs return the metrics definitions associated with this resource
func (rd ResourceDef) GetMetricDefs() []config.RextMetricDef {
	return rd.metrics
}

// GetOptions return key/value pairs for extra options
func (rd *ResourceDef) GetOptions() config.RextKeyValueStore {
	if rd.options == nil {
		rd.options = NewOptionsMap()
	}
	return rd.options
}

// Validate the resource, return true if any error is found
func (rd ResourceDef) Validate() bool {
	return config.ValidateResource(&rd)
}

// NewResourceDef create a new metric definition
func NewResourceDef(mType, resourceURI string, auth config.RextAuthDef, metrics []config.RextMetricDef, decoder config.RextDecoderDef, options config.RextKeyValueStore) config.RextResourceDef {
	return &ResourceDef{
		mType:       mType,
		resourceURI: resourceURI,
		auth:        auth,
		decoder:     decoder,
		metrics:     metrics,
		options:     options,
	}
}
