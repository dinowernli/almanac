package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	appenderFanout = 2 // The number of appenders each ingester talks to.
	httpPort       = 12345
)

var (
	conf = &config{
		smallChunkMaxEntries: 10,
		smallChunkSpreadMs:   5000,
		smallChunkMaxAgeMs:   3000,
	}

	appenderPorts = []int{5001, 5002, 5003, 5004, 5005}
)

func main() {
	logger := logrus.New()
	logger.Out = os.Stderr

	cluster, err := createCluster(logger, conf, appenderPorts, appenderFanout)
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
	logger.Infof("Started mixer at http://localhost:%d/mixer", httpPort)

	cluster.ingester.RegisterHttp(mux)
	logger.Infof("Started ingester at http://localhost:%d/ingester", httpPort)

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
}
