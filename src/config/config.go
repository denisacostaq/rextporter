package config

import (
	"errors"
)

var (
	// ErrKeyInvalidType for unexpected type
	ErrKeyInvalidType = errors.New("Unsupported type")
	// ErrKeyNotFound in key value store
	ErrKeyNotFound = errors.New("Missing key")
	// ErrKeyNotClonable in key value store
	ErrKeyNotClonable = errors.New("Impossible to obtain a copy of object")
	// ErrKeyConfigHaveSomeErrors for config validation
	ErrKeyConfigHaveSomeErrors = errors.New("Config have some errors")
	// ErrKeyEmptyValue values not allowed
	ErrKeyEmptyValue = errors.New("A required value is missed (empty or nil)")
	// ErrKeyDecodingFile can not parse or decode content
	ErrKeyDecodingFile = errors.New("Error decoding/parsing read content")
	// ErrKeyNotSuccessResponse can not get the data
	ErrKeyNotSuccessResponse = errors.New("Error getting data")
	// ErrKeyNotSupported operation not supported(may be implemented in future)
	ErrKeyNotSupported = errors.New("Unsupported operation")
)

const (
	// KeyMetricTypeCounter is the key you should define in the config file for counters.
	KeyMetricTypeCounter = "Counter"
	// KeyMetricTypeGauge is the key you should define in the config file for gauges.
	KeyMetricTypeGauge = "Gauge"
	// KeyMetricTypeHistogram is the key you should define in the config file for histograms.
	KeyMetricTypeHistogram = "Histogram"
	// KeyMetricTypeSummary is the key you should define in the config file for summaries.
	KeyMetricTypeSummary = "Summary"
)

const (
	// OptKeyRextResourceDefHTTPMethod key to define an http method inside a RextResourceDef
	OptKeyRextResourceDefHTTPMethod = "d43e326a-3e5d-462c-ad92-39dc2272f1d8"
	// OptKeyRextAuthDefTokenHeaderKey key to define a token header key inside a RextAuthDef
	OptKeyRextAuthDefTokenHeaderKey = "768772f5-cbe7-4a61-96ba-72ab99aede59" // nolint gosec
	// OptKeyRextAuthDefTokenKeyFromEndpoint key to define a token key from a response auth API inside a RextAuthDef
	OptKeyRextAuthDefTokenKeyFromEndpoint = "1cb99a48-c642-4234-af5e-7de88cb20271" // nolint gosec
	// OptKeyRextAuthDefTokenGenEndpoint key to define a token endpoint to get authenticated inside a RextAuthDef
	OptKeyRextAuthDefTokenGenEndpoint = "3a5e1d2f-53c0-4c47-b0cb-13a3190ce97f" // nolint gosec
	// OptKeyRextServiceDefJobName key to define the job name, it is mandatory for all services
	OptKeyRextServiceDefJobName = "555efe9a-fd0a-4f03-9724-fed758491e65"
	// OptKeyRextServiceDefInstanceName key to define a instance name for a service, it is mandatory for all services
	// a service can run in multiple nodes(physical or virtual), all these instances are mandatory, can be
	// for example 127.0.0.0:8080
	OptKeyRextServiceDefInstanceName = "0a12a60a-6ed4-400b-af78-2664d6588233"
	// OptKeyRextMetricDefHMetricBuckets key to hold the configured buckets inside a RextMetricDef if you are using
	// a histogram kind
	OptKeyRextMetricDefHMetricBuckets = "9983807d-13fe-4b1d-9363-4b844ea2f301"
)

const (
	// KeyLabelJob the label to map the job name for this system, can be a cluster
	KeyLabelJob = "job"
	// KeyLabelInstance is the label to map the instace in which this program is running as part of the job cluster
	KeyLabelInstance = "instance"
	// SystemProgramName is the name of this program
	SystemProgramName = "rextporter"
)

// RextRoot hold a service list whit their configurations info
type RextRoot interface {
	GetServices() []RextServiceDef
	AddService(RextServiceDef)
	Clone() (RextRoot, error)
	Validate() (hasError bool)
}

// RextServiceDef encapsulates all data for services
type RextServiceDef interface {
	SetBasePath(path string) // can be an http server base path, a filesystem directory ...
	GetBasePath() string
	// file | http | ftp
	GetProtocol() string
	SetProtocol(string) // TODO(denisacostaq@gmail.com): move this to set base path, and add a port too
	SetAuthForBaseURL(RextAuthDef)
	GetAuthForBaseURL() RextAuthDef
	AddResource(source RextResourceDef)
	AddResources(sources ...RextResourceDef)
	GetResources() []RextResourceDef
	GetOptions() RextKeyValueStore
	Clone() (RextServiceDef, error)
	Validate() (hasError bool)
}

// RextResourceDef for retrieving raw data
type RextResourceDef interface {
	// GetResourcePATH should be used in the context of a service, so the service base path information
	// have to be passed to this method, it returns a resource url = base_path + uri
	GetResourcePATH(basePath string) string

	// GetAuth should be used in the context of a service, so the service auth information
	// have to be passed to this method(can be null if not auth is required for major service calls)
	// it returns the resource auth info or the general auth for service if the resource have not a specific
	// one.
	GetAuth(defAuth RextAuthDef) (auth RextAuthDef)

	// SetResourceURI set the path where live the resource inside a service, see examples below
	// http -> /api/v1/network/connections | /api/v1/health
	// file -> /path/to/a/file | /proc/$(pidof qtcreator)/status
	// the retrieved resource can be a json file, a xml, a plain text,  a .rar ...
	SetResourceURI(string)

	// SetAuth set a specific auth info for the resource if required, for example
	// in a web server different resource path can have different different auth strategics|info,
	// in a filesystem some special files may require root(admin) access
	SetAuth(RextAuthDef)

	// GetDecoder return a decoder to parse the resource and get the info
	GetDecoder() RextDecoderDef

	// SetDecoder set a decoder to parse the resource and get the info
	SetDecoder(RextDecoderDef)

	// AddMetricDef set a metric definition for this resource path
	AddMetricDef(RextMetricDef)
	GetMetricDefs() []RextMetricDef

	SetType(string)  // TODO(denisacostaq@gmail.com): remove this
	GetType() string // TODO(denisacostaq@gmail.com): remove this
	GetOptions() RextKeyValueStore
	Clone() (RextResourceDef, error)
	Validate() (hasError bool)
}

// RextDecoderDef allow you to decode a resource from different formats
type RextDecoderDef interface {
	// GetType return some kind of "encoding" like: json, xml, ini, plain_text, prometheus_exposed_metrics,
	// .rar(even encrypted)
	GetType() string

	// GetOptions return additional options for example if the retrieved content is encripted, get info
	// about the algorithm, the key, and so on...
	GetOptions() RextKeyValueStore
	Clone() (RextDecoderDef, error)
	Validate() (hasError bool)
}

const (
	// RextNodeSolverTypeJSONPath var name to use node solver of json kind
	RextNodeSolverTypeJSONPath = "jsonPath"
)

// RextNodeSolver help you to get raw data(sample/s) to create a metric from a specific path inside a
// retrieved resource
type RextNodeSolver interface {
	// GetType return the strategy to find the data, it can be: jpath, xpath, .ini, plain_text, .tar.gz
	// it is different to RextDecoderDef.type in the sense of a decoder can work over a binary encoded
	// content and after, the node solver over a .rar
	GetType() string

	// GetNodePath return the path where you can find the value, it depends on the type, see some examples below:
	// "json" -> "/blockchain/head/seq" | "/blockchain/head/fee"
	// "xml" -> "/blockchain/head/seq" | "/blockchain/head/fee"
	// "ini" -> "key_name"
	// "plain_text" -> line number
	// "directory" -> file_path
	// ".rar" -> file_path | file_path + jpath for the specific file | file_path + key(.ini) for the specific file
	GetNodePath() string
	SetNodePath(string)

	// GetOptions return additional information for more complex data structures, like for example in the
	// .rar example above
	GetOptions() RextKeyValueStore
	Clone() (RextNodeSolver, error)
	Validate() (hasError bool)
}

// RextMetricDef contains the metadata associated to the metrics
type RextMetricDef interface {
	// GetMetricName return the metric name
	GetMetricName() string
	// GetMetricType return the metric type
	GetMetricType() string
	// GetMetricDescription return the metric description
	GetMetricDescription() string
	// GetLabels return the labels in which the metrics should be mapped in
	GetLabels() []RextLabelDef
	// GetNodeSolver return a solver able to get the metric sample/s
	GetNodeSolver() RextNodeSolver
	SetNodeSolver(RextNodeSolver)
	SetMetricName(string)
	SetMetricType(string)
	SetMetricDescription(string)
	AddLabel(RextLabelDef)
	GetOptions() RextKeyValueStore
	Clone() (RextMetricDef, error)
	Validate() (hasError bool)
}

// RextLabelDef define a label name and the way to get the value for metrics vec
type RextLabelDef interface {
	SetName(string)
	// GetName return the metric name
	GetName() string
	SetNodeSolver(RextNodeSolver)
	// GetNodeSolver return the solver able to get the metric value
	GetNodeSolver() RextNodeSolver
	Clone() (RextLabelDef, error)
	Validate() (hasError bool)
}

// AuthTypeCSRF define a const name for auth of type CSRF
const AuthTypeCSRF = "CSRF"

// RextAuthDef can store information about authentication requirements, how and where you can autenticate,
// using what values, all this info is stored inside a RextAuthDef
type RextAuthDef interface {
	// SetAuthType set the auth type
	SetAuthType(string)
	// GetAuthType return the auth type
	GetAuthType() string
	GetOptions() RextKeyValueStore
	Clone() (RextAuthDef, error)
	Validate() (hasError bool)
}

// RextKeyValueStore providing access to object settings, you give a key with a value(can be a string or
// a interface{}) for store this value and later you can get back this value trough the original key.
type RextKeyValueStore interface {
	GetString(key string) (string, error)
	SetString(key string, value string) (bool, error)
	GetObject(key string) (interface{}, error)
	SetObject(key string, value interface{}) (bool, error)
	GetKeys() []string
	Clone() (RextKeyValueStore, error)
}
