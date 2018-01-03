package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

const (
	entriesPerChunk = 10 // The max number of entries in a single chunk.
	numAppenders    = 5  // The number of appenders in the system.
	appenderFanout  = 2  // The number of appender each ingester talks to.

	httpPort     = 12345
	grpcBasePort = 51000
)

func main() {
	cluster, err := createCluster(grpcBasePort, numAppenders, entriesPerChunk, appenderFanout)
	if err != nil {
		panic(err)
	}

	ingestRequest1, err := ingestRequest(&entry{Message: "foo", TimestampMs: 5000})
	if err != nil {
		panic(err)
	}

	ingestRequest2, err := ingestRequest(&entry{Message: "foo", TimestampMs: 5007})
	if err != nil {
		panic(err)
	}

	_, err = cluster.ingester.Ingest(context.Background(), ingestRequest1)
	if err != nil {
		panic(err)
	}

	_, err = cluster.ingester.Ingest(context.Background(), ingestRequest2)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	cluster.mixer.RegisterHttp(mux)

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
}
