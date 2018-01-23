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

	flagAppenderPorts = kingpin.Flag("appender_ports", "Which ports to run appenders on").Default("5001", "5002", "5003", "5004", "5005").Ints()
	flagHttpPort      = kingpin.Flag("http_port", "which port to run the http server on").Default("12345").Int()

	flagIngestFanout         = kingpin.Flag("ingest_fanout", "How many appenders to send each ingested entry to").Default("2").Int()
	flagSmallChunkMaxEntries = kingpin.Flag("small_chunk_max_entries", "The maximum number of entries in a small chunk").Default("10").Int()
	flagSmallChunkMaxSpread  = kingpin.Flag("small_chunk_max_spread", "The maximum spread of a small chunk").Default("5s").Duration()
	flagSmallChunkMaxAge     = kingpin.Flag("small_chunk_max_age", "The maximum time a small chunk can stay open").Default("3s").Duration()
	flagBigChunkMaxSpread    = kingpin.Flag("big_chunk_max_spread", "The maximum spread of a big chunk").Default("12h").Duration()

	flagJanitorCompactionInterval = kingpin.Flag("janitor_compaction_interval", "How frequently the janitor runs compactions").Default("10s").Duration()
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kingpin.Parse()
	logger := logrus.New()
	logger.Out = os.Stderr

	conf := &cluster.Config{
		SmallChunkMaxEntries: *flagSmallChunkMaxEntries,
		SmallChunkSpread:     *flagSmallChunkMaxSpread,
		SmallChunkMaxAge:     *flagSmallChunkMaxAge,
		BigChunkMaxSpread:    *flagBigChunkMaxSpread,

		JanitorCompactionInterval: *flagJanitorCompactionInterval,

		StorageType: *flagStorageType,
		GcsBucket:   *flagGcsBucket,
	}

	cluster, err := cluster.CreateCluster(ctx, logger, conf, *flagAppenderPorts, *flagIngestFanout)
	if err != nil {
		panic(err)
	}

	ingestRequest1 := &pb_almanac.IngestRequest{EntryJson: `{ "message": "foo", "timestamp_ms": 5000 }`}
	_, err = cluster.Ingester.Ingest(context.Background(), ingestRequest1)
	if err != nil {
		panic(err)
	}

	ingestRequest2 := &pb_almanac.IngestRequest{EntryJson: `{ "message": "foo", "timestamp_ms": 5007 }`}
	_, err = cluster.Ingester.Ingest(context.Background(), ingestRequest2)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	cluster.Mixer.RegisterHttp(mux)
	logger.Infof("Started mixer at http://localhost:%d/mixer", *flagHttpPort)

	cluster.Ingester.RegisterHttp(mux)
	logger.Infof("Started ingester at http://localhost:%d/ingester", *flagHttpPort)

	http.ListenAndServe(fmt.Sprintf(":%d", *flagHttpPort), mux)
}
