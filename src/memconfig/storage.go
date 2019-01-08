package memconfig

import (
	"github.com/simelo/rextporter/src/config"
)

// OptionsMap in-memory key value store
type OptionsMap map[string]interface{}

// NewOptionsMap creates a new instance
func NewOptionsMap() (m OptionsMap) {
	m = make(OptionsMap)
	return
}

// GetString return the string value for key
func (m OptionsMap) GetString(key string) (string, error) {
	var err error
	var val interface{}
	if val, err = m.GetObject(key); err == nil {
		strVal, okStrVal := val.(string)
		if okStrVal {
			return strVal, nil
		}
		return "", config.ErrKeyInvalidType
	}
	return "", err
}

// SetString set a string value for key
func (m OptionsMap) SetString(key string, value string) (exists bool, err error) {
	return m.SetObject(key, value)
}

// GetObject return a saved object
func (m OptionsMap) GetObject(key string) (interface{}, error) {
	if val, hasKey := m[key]; hasKey {
		return val, nil
	}
	return "", config.ErrKeyNotFound
}

// SetObject save an general object
func (m OptionsMap) SetObject(key string, value interface{}) (exists bool, err error) {
	err = nil
	_, exists = m[key]
	m[key] = value
	return
}

// GetKeys return all the saved keys
func (m OptionsMap) GetKeys() (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

// Clone make a deep copy of the storage
func (m OptionsMap) Clone() (config.RextKeyValueStore, error) {
	clone := NewOptionsMap()
	for k := range m {
		clone[k] = m[k]
	}
	return clone, nil
}
