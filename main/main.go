package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	numAppenders   = 5 // The number of appenders in the system.
	appenderFanout = 2 // The number of appender each ingester talks to.

	httpPort     = 12345
	grpcBasePort = 51000
)

var (
	conf = &config{
		smallChunkMaxEntries: 10,
		smallChunkSpreadMs:   5000,
		smallChunkMaxAgeMs:   3000,
	}
)

func main() {
	logger := logrus.New()
	logger.Out = os.Stderr

	cluster, err := createCluster(logger, conf, grpcBasePort, numAppenders, appenderFanout)
	if err != nil {
		panic(err)
	}

	ingestRequest1, err := newIngestRequest(&entry{Message: "foo", TimestampMs: 5000})
	if err != nil {
		panic(err)
	}

	ingestRequest2, err := newIngestRequest(&entry{Message: "foo", TimestampMs: 5007})
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
	cluster.ingester.RegisterHttp(mux)

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
}
