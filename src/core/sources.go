package core

// RextDataSource for retrieving raw data
type RextDataSource interface {
	SetBaseURL(url string)
	GetMethod() string
	SetMethod(string)
	GetResourceLocation() string
	SetResourceLocation(string) error
	// FIXME: Scraping orthogonal to data source?
	ActivateScraper(RextMetricsExtractor) error
	GetOptions() RextKeyValueStore
}
