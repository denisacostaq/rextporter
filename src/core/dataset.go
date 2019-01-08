package core

import (
	"errors"
)

var (
	// ErrInvalidType for unecpected type
	ErrInvalidType = errors.New("Unsupported type")
	// ErrKeyNotFound in key value store
	ErrKeyNotFound = errors.New("Missing key")
	// ErrNotClonable in key value store
	ErrNotClonable = errors.New("Impossible to obtain a copy of object")
)

// RextKeyValueStore providing access to object settings
type RextKeyValueStore interface {
	GetString(key string) (string, error)
	SetString(key string, value string) (bool, error)
	GetObject(key string) (interface{}, error)
	SetObject(key string, value interface{}) (bool, error)
	GetKeys() []string
	Clone() (RextKeyValueStore, error)
}

// RextAuth implements an authentication strategies
type RextAuth interface {
	GetAuthType() string
	GetOptions() RextKeyValueStore
}

// RextMetric provides access to values measured for a given metric
type RextMetric interface {
	GetMetadata() RextMetricDef
	// TODO: Methods to retrieve values measured for a given metric
}
