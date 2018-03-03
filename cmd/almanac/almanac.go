package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dinowernli/almanac/pkg/cluster"
	"github.com/dinowernli/almanac/pkg/storage"
	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/alecthomas/kingpin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	metricsHttpPath = "/metrics"
)

var (
	flagStorageType = kingpin.Flag("storage", "Which kind of storage to use").Default(storage.StorageTypeMemory).Enum(storage.StorageTypeMemory, storage.StorageTypeDisk, storage.StorageTypeGcs)
	flagGcsBucket   = kingpin.Flag("storage.gcs.bucket", "Which gcs bucket to use for storage").Default("almanac-dev").String()
	flagDiskPath    = kingpin.Flag("storage.disk.path", "An existing empty directory to use as root for storage").Default("/tmp/almanac-dev").String()

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
		DiskPath:    *flagDiskPath,
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
	mux.Handle(metricsHttpPath, prometheus.InstrumentHandler(metricsHttpPath, promhttp.Handler()))
	logger.Infof("Metrics at %s", formatLink(metricsHttpPath))

	cluster.Mixer.RegisterHttp(mux)
	logger.Infof("Mixer at %s", formatLink("/mixer"))

	cluster.Ingester.RegisterHttp(mux)
	logger.Infof("Ingester at %s", formatLink("/ingester"))

	http.ListenAndServe(fmt.Sprintf(":%d", *flagHttpPort), mux)
}

func formatLink(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", *flagHttpPort, path)
}
