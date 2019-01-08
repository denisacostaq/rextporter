package wip

import "github.com/simelo/rextporter/src/core"

// RextMetric provides access to values measured for a given metric
type RextMetric interface {
	GetMetadata() core.RextMetricDef
	// TODO: Methods to retrieve values measured for a given metric
}
