package util

import (
	"net/url"

	"github.com/simelo/rextporter/src/config"
)

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

// IsValidURL tests a string to determine if it is a valid URL or not.
func IsValidURL(toTest string) bool {
	if _, err := url.ParseRequestURI(toTest); err != nil {
		return false
	}
	return true
}
