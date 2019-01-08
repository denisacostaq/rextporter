package core

// RextServiceScraper encapsulates all data and logic for scraping services
type RextServiceScraper interface {
	AddAuthStrategy(auth RextAuth, name string)
	AddSource(source RextDataSource)
	AddSources(sources ...RextDataSource)
	GetOptions() RextKeyValueStore
}

// RextMetricScraper extract metrics from raw data based on scraping rules
type RextMetricsExtractor interface {
	Apply(rule RextMetricDef) (RextMetricDef, error)
	ApplyMany(rules []RextMetricDef) ([]RextMetricDef, error)
	ExtractMetrics(target interface{}) ([]RextMetric, error)
	GetOptions() RextKeyValueStore
}

// RextMetricDef contains the metadata associated to performance metrics
type RextMetricDef interface {
	GetMetricName() string
	GetMetricType() string
	GetMetricDescription() string
	GetMetricLabels() []string
	SetMetricName(string)
	SetMetricType(string)
	SetMetricDescription(string)
	SetMetricLabels([]string)
	GetOptions() RextKeyValueStore
}
