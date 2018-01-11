package main

import (
	"fmt"
	"net/http"
	"os"

	"dinowernli.me/almanac/pkg/cluster"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var (
	flagStorageType = kingpin.Flag("storage", "Which kind of storage to use").Default("memory").Enum("memory", "gcs")
	flagGcsBucket   = kingpin.Flag("gcs.bucket", "Which gcs bucket to use for storage").Default("almanac-dev").String()

	appenderPorts = []int{5001, 5002, 5003, 5004, 5005}
)

func main() {
	kingpin.Parse()
	logger := logrus.New()
	logger.Out = os.Stderr

	appenderFanout := 2
	httpPort := 12345

	conf := &cluster.Config{
		SmallChunkMaxEntries: 10,
		SmallChunkSpreadMs:   5000,
		SmallChunkMaxAgeMs:   3000,

		StorageType: *flagStorageType,
		GcsBucket:   *flagGcsBucket,
	}

	cluster, err := cluster.CreateCluster(logger, conf, appenderPorts, appenderFanout)
	if err != nil {
		panic(err)
	}

	ingestRequest1 := newIngestRequest("{ \"message\": \"foo\", \"timestamp_ms\": 5000 }")
	_, err = cluster.Ingester.Ingest(context.Background(), ingestRequest1)
	if err != nil {
		panic(err)
	}

	ingestRequest2 := newIngestRequest("{ \"message\": \"foo\", \"timestamp_ms\": 5007 }")
	_, err = cluster.Ingester.Ingest(context.Background(), ingestRequest2)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	cluster.Mixer.RegisterHttp(mux)
	logger.Infof("Started mixer at http://localhost:%d/mixer", httpPort)

	cluster.Ingester.RegisterHttp(mux)
	logger.Infof("Started ingester at http://localhost:%d/ingester", httpPort)

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux)
}

func newIngestRequest(json string) *pb_almanac.IngestRequest {
	return &pb_almanac.IngestRequest{EntryJson: json}
}
