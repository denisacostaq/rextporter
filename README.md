
# Rextporter

[![Build Status](https://travis-ci.org/simelo/rextporter.svg?branch=develop)](https://travis-ci.org/simelo/rextporter)

This is the executable entry point for the server.
- [Install](#install)
- [Run](#run)
- [Config file](#config-file)

## Install

```bash
$ cd $GOPATH/src/github.com/simelo/rextporter/cmd/rextporter
$ go install ./...
```

## Run

You can run the program (`rextporter`, make sure you have it accessible trough your `PATH` env variable) by calling it in the console and you have the following parameters options.

 - `-config` Metrics main config file path. This flag is required.
 - `-handler` Handler to expose metric. (default "/metrics").
 - `-port` Listen port. (default 8080)

### Config file

Trough program parameter you should refer to main config path.

You should have at least 6 config files:

- Main config (general definitions, like for example, load the service config from file and use "this" path).

- Services config (services definitions), the path should be maped from the main config.

- Metrics for service (metrics definitions) config, the path should be maped from the main config.

- Resources paths for service (resources paths definitions) config, the path should be maped from the main config.

- `ServiceName+Metrics.toml` define the metrics of a giving service (`ServiceName`), the path should be mapped from metrics for service. You can have multiple ServiceNameMetric(`skycoinMetric.toml` for instance), depending on the number of service and how they are mapped with metrics.

- `ServiceName+ResourcePath.toml` define the available resource in a giving service (`ServiceName`), the path should be mapped from resources paths for service. You can have multiple ServiceNameResourcesPaths (`skycoinResourcesPaths.toml` for instance), depending on the number of service and how they are mapped with metrics.

Example main configuration file:
```toml
servicesConfigPath = "tomlconfig/services.toml"
metricsForServicesConfigPath = "tomlconfig/metricsForServices.toml"
resourcePathsForServicesConfPath = "tomlconfig/resourcePathsForServices.toml"
```

Example services configuration file:
```toml
# Services configuration.
[[services]]
	name = "skycoin"
	protocol = "http"
	port = 6420
	authType = "CSRF"
	tokenHeaderKey = "X-CSRF-Token"
	genTokenEndpoint = "/api/v1/csrf"
	tokenKeyFromEndpoint = "csrf_token"

	[services.location]
		location = "localhost"

```

Example metrics for service configuration file:
```toml
metricPathsForServicesConfig = [
	{ skycoin = "tomlconfig/skycoinMetrics.toml" },
]
```

Example resources paths for services configuration file:
```toml
resourcePathsForServicesConfig = [
	{ skycoin = "tomlconfig/skycoinResourcesPaths.toml" },
]
```

Example metrics(for skycoin in this case) configuration file.
```toml
[[metrics]]
	name = "health_seq"
	path = "/blockchain/head/seq"

	[metrics.options]
		type = "Counter"
		description = "Seq value from endpoint /api/v1/health, json node blockchain -> head -> seq"
```

Example resources paths(for skycoin in this case) configuration file.
```toml
[[ResourcePaths]]
	Name = "health"
	Path = "/api/v1/health"
	PathType = "rest_api"
	nodeSolverType = "jsonPath"
	MetricNames = ["health_seq"]
```
The `MetricNames` allow you to enable only a subset of all the available metrics for this resource path.

Example gauge vector metric configuration.
```toml
[[metrics]]
	name = "connections_highest_by_address"
	path = "/connections/height"

	[metrics.options]
		type = "Gauge"
		description = "Value from endpoint /api/v1/network/connections, json node connections -> highest" 
		[[metrics.options.labels]]
			name = "Address"
			path = "/connections/address"
```

Example histogram metric configuration:
```toml
[[metrics]]
	name = "connections_burn_factor_hist"
	path = "/connections/unconfirmed_verify_transaction/burn_factor"

	[metrics.options]
		type = "Histogram"
		description = "Burn factor histogram across connections"
	
	[metrics.histogramOptions]
		buckets = [1, 2, 3]
```

A full example configuration for skycoin can be found in the [integration tests folder](https://github.com/simelo/rextporter/tree/master/test/integration/skycoin/tomlconfig).
