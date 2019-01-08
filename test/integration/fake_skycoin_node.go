package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/simelo/rextporter/src/util/file"
	"github.com/simelo/rextporter/test/util/testrand"
	log "github.com/sirupsen/logrus"
)

func apiHealthHandler(w http.ResponseWriter, r *http.Request) {
	const jsonHealthResponse = `
{
    "blockchain": {
        "head": {
            "seq": 58894,
            "block_hash": "3961bea8c4ab45d658ae42effd4caf36b81709dc52a5708fdd4c8eb1b199a1f6",
            "previous_block_hash": "8eca94e7597b87c8587286b66a6b409f6b4bf288a381a56d7fde3594e319c38a",
            "timestamp": 1537581604,
            "fee": 485194,
            "version": 0,
            "tx_body_hash": "c03c0dd28841d5aa87ce4e692ec8adde923799146ec5504e17ac0c95036362dd",
            "ux_hash": "f7d30ecb49f132283862ad58f691e8747894c9fc241cb3a864fc15bd3e2c83d3"
        },
        "unspents": 38171,
        "unconfirmed": 1,
        "time_since_last_block": "4m46s"
    },
    "version": {
        "version": "0.24.1",
        "commit": "8798b5ee43c7ce43b9b75d57a1a6cd2c1295cd1e",
        "branch": "develop"
    },
    "open_connections": 8,
    "uptime": "6m30.629057248s",
    "csrf_enabled": true,
    "csp_enabled": true,
    "wallet_api_enabled": true,
    "gui_enabled": true,
    "unversioned_api_enabled": false,
    "json_rpc_enabled": false
}
`
	if _, err := w.Write([]byte(jsonHealthResponse)); err != nil {
		log.WithError(err).Errorln("unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func apiConnectionsHandlerHandler(w http.ResponseWriter, r *http.Request) {
	const jsonConnectionsResponse = `
	{
    "connections": [
        {
            "id": 99107,
            "address": "139.162.161.41:20002",
            "last_sent": 1520675750,
            "last_received": 1520675750,
            "connected_at": 1520675500,
            "outgoing": false,
            "state": "introduced",
            "mirror": 1338939619,
            "listen_port": 20002,
            "height": 180,
            "user_agent": "skycoin:0.25.0",
		    "is_trusted_peer": true,
		    "unconfirmed_verify_transaction": {
		        "burn_factor": 2,
		        "max_transaction_size": 32768,
		        "max_decimals": 3
		    }
        },
        {
            "id": 109548,
            "address": "176.9.84.75:6000",
            "last_sent": 1520675751,
            "last_received": 1520675751,
            "connected_at": 1520675751,
            "state": "connected",
            "outgoing": true,
            "mirror": 0,
            "listen_port": 6000,
            "height": 0,
            "user_agent": "",
		    "is_trusted_peer": true,
		    "unconfirmed_verify_transaction": {
		        "burn_factor": 0,
		        "max_transaction_size": 0,
		        "max_decimals": 0
		    }
        },
        {
            "id": 99115,
            "address": "185.120.34.60:6000",
            "last_sent": 1520675754,
            "last_received": 1520675754,
            "connected_at": 1520673013,
            "outgoing": false,
            "state": "introduced",
            "mirror": 1931713869,
            "listen_port": 6000,
            "height": 180,
            "user_agent": "",
		    "is_trusted_peer": true,
		    "unconfirmed_verify_transaction": {
		        "burn_factor": 0,
		        "max_transaction_size": 0,
		        "max_decimals": 0
		    }
        }
    ]
}
`
	if _, err := w.Write([]byte(jsonConnectionsResponse)); err != nil {
		log.WithError(err).Errorln("unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func exposedMetricHandler(w http.ResponseWriter, r *http.Request) {
	const exposedMetricsResponse = `
# HELP go_gc_duration_seconds1a18ac9b29c6 A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds1a18ac9b29c6 summary
go_gc_duration_seconds1a18ac9b29c6{quantile="0"} 0
go_gc_duration_seconds1a18ac9b29c6{quantile="0.25"} 0
go_gc_duration_seconds1a18ac9b29c6{quantile="0.5"} 0
go_gc_duration_seconds1a18ac9b29c6{quantile="0.75"} 0
go_gc_duration_seconds1a18ac9b29c6{quantile="1"} 0
go_gc_duration_seconds1a18ac9b29c6_sum 0
go_gc_duration_seconds1a18ac9b29c6_count 0
# HELP go_goroutines1a18ac9b29c6 Number of goroutines that currently exist.
# TYPE go_goroutines1a18ac9b29c6 gauge
go_goroutines1a18ac9b29c6 7
# HELP go_memstats_mallocs_total1a18ac9b29c6 Total number of mallocs.
# TYPE go_memstats_mallocs_total1a18ac9b29c6 counter
go_memstats_mallocs_total1a18ac9b29c6 9049
# HELP promhttp_metric_handler_requests_total1a18ac9b29c6 Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total1a18ac9b29c6 counter
promhttp_metric_handler_requests_total1a18ac9b29c6{code="200"} 0
promhttp_metric_handler_requests_total1a18ac9b29c6{code="500"} 0
promhttp_metric_handler_requests_total1a18ac9b29c6{code="503"} 0
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 17
`
	if _, err := w.Write([]byte(exposedMetricsResponse)); err != nil {
		log.WithError(err).Errorln("unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func exposedAFewMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics = `
# HELP main_seq Says if the same name metric(skycoin_wallet2_seq2) was success updated, 1 for ok, 0 for failed.
# TYPE main_seq gauge
main_seq 13
# HELP main_seq_up Says if the same name metric(skycoin_wallet2_seq2) was success updated, 1 for ok, 0 for failed.
# TYPE main_seq_up gauge
main_seq_up 0
# HELP seq Says if the same name metric(skycoin_wallet2_seq2) was success updated, 1 for ok, 0 for failed.
# TYPE seq gauge
seq 32
# HELP seq_up Says if the same name metric(skycoin_wallet2_seq2) was success updated, 1 for ok, 0 for failed.
# TYPE seq_up gauge
seq_up 0
`
	if _, err := w.Write([]byte(metrics)); err != nil {
		log.WithError(err).Errorln("unable to write response")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func writeListenPortInFile(port uint16) (err error) {
	var path string
	path, err = testrand.FilePathToSharePort()
	if err != nil {
		return err
	}
	if !file.ExistFile(path) {
		var file, err = os.Create(path)
		if err != nil {
			log.WithError(err).Errorln("error creating file")
			return err
		}
		defer file.Close()
	}
	var file *os.File
	file, err = os.OpenFile(path, os.O_WRONLY, 0400)
	if err != nil {
		log.WithError(err).Errorln("error opening file")
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("%d", port))
	if err != nil {
		log.WithError(err).Errorln("error writing file")
		return err
	}
	err = file.Sync()
	if err != nil {
		log.WithError(err).Errorln("error flushing file")
		return err
	}
	return err
}

func main() {
	var fakeNodePort = testrand.RandomPort()
	if err := writeListenPortInFile(fakeNodePort); err != nil {
		log.Fatal(err)
	}
	log.WithField("port", fakeNodePort).Infoln("starting fake server")
	http.HandleFunc("/api/v1/health", apiHealthHandler)
	http.HandleFunc("/api/v1/network/connections", apiConnectionsHandlerHandler)
	http.HandleFunc("/metrics2", exposedMetricHandler)
	http.HandleFunc("/a_few_metrics", exposedAFewMetrics)
	log.WithError(http.ListenAndServe(fmt.Sprintf(":%d", fakeNodePort), nil)).Fatalln("server fail")
}
