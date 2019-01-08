package core

// RextEnv encapsulates an implementation of the rextporter top-level environment
type RextEnv interface {
	NewServiceScraper() (RextServiceScraper, error)
	NewAuthStrategy(authtype string, options RextKeyValueStore) (RextAuth, error)
	NewMetricsExtractor(scraperType string, options RextKeyValueStore, metrics []RextMetricDef) (RextMetricsExtractor, error)
	NewMetricsDatasource(srcType string) (RextDataSource, error)
	RegisterScraperForServices(s RextServiceScraper, services ...string) error
	RegisterScraperForStacks(s RextServiceScraper, stacks ...string) error
	GetOptions() RextKeyValueStore
}
