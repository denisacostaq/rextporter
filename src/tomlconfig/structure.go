package tomlconfig

// RootConfig is the top level node for the config tree, it has a list of services
type RootConfig struct {
	Services []Service
}

// Service is a concept to grab information about a datasource, for example:
// where is it http://localhost:1234 (Protocol + Location + : + Port + BasePath)
// what is the filesystem path(in case of file protocol)?
type Service struct {
	Name string
	// Protocol is file, http, https
	Protocol string
	Port     uint16
	// FIXME(denisacostaq@gmial.com): use this base path?
	BasePath             string
	AuthType             string
	TokenHeaderKey       string
	GenTokenEndpoint     string
	TokenKeyFromEndpoint string
	Location             Server
	ResourcePaths        ResourcePathTemplate
	Metrics              MetricsTemplate
}

// MetricsTemplate is a list of metrics definition, ready to be applied
// multiple times to different services, to apply the same metric template to different
// services with different metrics check out how ResourcePathTemplate can
// define a subset of the metrics in a template
type MetricsTemplate []Metric

// Metric describe a metadata about how to load real exposed metrics
type Metric struct {
	Name             string           `json:"name"`
	Path             string           `json:"path,omitempty"`
	Options          MetricOptions    `json:"options"`
	HistogramOptions HistogramOptions `json:"histogram_options"`
}

// MetricOptions keep information you about the metric, mostly the type(Counter, Gauge, Summary, and Histogram)
type MetricOptions struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Labels      []Label
}

// Label to create metrics grouping by json path value, for example:
// {Name: "color", "/properties/color"}
type Label struct {
	// Name the name of the label, different values can be assigned to it
	Name string
	// Path the json path from where you need to extract the label
	Path string
}

// HistogramOptions allows you to define the histogram is buckets.
type HistogramOptions struct {
	Buckets []float64 `json:"buckets"`

	// ExponentialBuckets is a len three array where:
	// - The first value is the low bound start bucket.
	// - The second vale is the growing factor.
	// - The three one is the buckets amount.
	ExponentialBuckets []float64 `json:"exponential_buckets"`
}

// ResourcePath define a node solver type for a giving resource inside a service
// in adition, have a metric names list, it can work like a filter
// to work over a subset of the defined metrics for a resource
// TODO(denisacostaq@gmail.com) this filter should work for metrics fordwader too
type ResourcePath struct {
	Name           string
	PathType       string
	Path           string
	NodeSolverType string
	HTTPMethod     string
	// MetricNames TODO(denisacostaq@gmail.com): trying to define filtered metric can introduce
	// some redundancy because the other fields
	MetricNames []string
}

// ResourcePathTemplate can be used to define subset of metrics from MetricsTemplate in a giving
// service
type ResourcePathTemplate []ResourcePath

// Server the server where is running the service
type Server struct {
	// Location should have the ip or URL.
	Location string `json:"location"`
}
