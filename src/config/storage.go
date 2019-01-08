package config

import (
	"github.com/simelo/rextporter/src/core"
)

// OptionsMap in-memory key value store
type OptionsMap map[string]string

// NewOptionsMap creates a new instance
func NewOptionsMap() (m OptionsMap) {
	m = OptionsMap(make(map[string]string, 0))
	return
}

// GetString value for key
func (m OptionsMap) GetString(key string) (string, error) {
	if val, hasKey := m[key]; hasKey {
		return val, nil
	}
	return "", core.ErrKeyNotFound
}

// SetString value for key
func (m OptionsMap) SetString(key string, value string) (exists bool, err error) {
	err = nil
	_, exists = m[key]
	m[key] = value
	return
}

// GetObject equivalent to GetString
func (m OptionsMap) GetObject(key string) (interface{}, error) {
	return m.GetString(key)
}

// SetObject only strings supported
func (m OptionsMap) SetObject(key string, value interface{}) (bool, error) {
	if s, isStr := value.(string); isStr {
		return m.SetString(key, s)
	}
	return false, core.ErrInvalidType
}

func (m OptionsMap) GetKeys() (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

func (m OptionsMap) Clone() (core.RextKeyValueStore, error) {
	clone := NewOptionsMap()
	for k := range m {
		clone[k] = m[k]
	}
	return clone, nil
}
