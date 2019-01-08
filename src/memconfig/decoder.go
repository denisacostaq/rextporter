package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// Decoder implements the interface config.RextMetricDef
type Decoder struct {
	mType   string
	options config.RextKeyValueStore
}

// Clone make a deep copy of Decoder or return an error if any
func (d Decoder) Clone() (cD config.RextDecoderDef, err error) {
	var cOpts config.RextKeyValueStore
	if cOpts, err = d.GetOptions().Clone(); err != nil {
		log.WithError(err).Errorln("can not clone options in decoder")
		return cD, err
	}
	cD = NewDecoder(d.mType, cOpts)
	return cD, err
}

// GetType return the decoder type
func (d Decoder) GetType() string {
	return d.mType
}

// GetOptions return key/value pairs for extra options
func (d *Decoder) GetOptions() config.RextKeyValueStore {
	if d.options == nil {
		d.options = NewOptionsMap()
	}
	return d.options
}

// Validate the decoder, return true if any error is found
func (d Decoder) Validate() bool {
	return config.ValidateDecoder(&d)
}

// NewDecoder create a new decoder
func NewDecoder(mType string, options config.RextKeyValueStore) *Decoder {
	return &Decoder{
		mType:   mType,
		options: options,
	}
}
