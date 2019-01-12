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

// MergeStoresInplace to update key / values in destination with those in source
func MergeStoresInplace(dst, src config.RextKeyValueStore) (err error) {
	var value interface{}
	err = nil
	for _, k := range src.GetKeys() {
		if value, err = src.GetObject(k); err == nil {
			if _, err = dst.SetObject(k, value); err != nil {
				return
			}
		} else {
			return
		}
	}
	return
}

// MergeStoresInANewOne create a new storage with the content of src1 and src2 merged
func MergeStoresInANewOne(src1, src2 config.RextKeyValueStore) (res config.RextKeyValueStore, err error) {
	res = NewOptionsMap()
	if err := MergeStoresInplace(res, src1); err != nil {
		return res, err
	}
	if err := MergeStoresInplace(res, src2); err != nil {
		return res, err
	}
	return res, nil
}
