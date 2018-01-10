package main

import (
	"encoding/json"
	"fmt"
	"net"

	"dinowernli.me/almanac/pkg/service/appender"
	dc "dinowernli.me/almanac/pkg/service/discovery"
	in "dinowernli.me/almanac/pkg/service/ingester"
	mx "dinowernli.me/almanac/pkg/service/mixer"
	st "dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	storageTypeMemory = "memory"
	storageTypeGcs    = "gcs"
)

// entry represents a log entry as would be supplied by a user of the system.
type entry struct {
	Message     string `json:"message"`
	Logger      string `json:"logger"`
	TimestampMs int64  `json:"timestamp_ms"`
}

// config holds a few configurable values defining the behavior of the system.
type config struct {
	smallChunkMaxEntries int
	smallChunkSpreadMs   int64
	smallChunkMaxAgeMs   int64

	storageType string
	gcsBucket   string
}

// localCluster holds a test setup ready to use for testing.
type localCluster struct {
	mixer    *mx.Mixer
	ingester *in.Ingester

	appenders []*appender.Appender
	storage   *st.Storage
	discovery *dc.Discovery

	servers []*grpc.Server
}

// createCluster sets up a test cluster, including all services required to run the system.
func createCluster(logger *logrus.Logger, config *config, appenderPorts []int, appenderFanout int) (*localCluster, error) {
	var err error
	var storage *st.Storage
	if config.storageType == storageTypeMemory {
		storage = st.NewInMemoryStorage()
	} else if config.storageType == storageTypeGcs {
		storage, err = st.NewGcsStorage(config.gcsBucket)
		if err != nil {
			return nil, fmt.Errorf("unable to create gcs storage: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unrecognized storage type: %s", config.storageType)
	}

	appenders := []*appender.Appender{}
	servers := []*grpc.Server{}
	appenderAddresses := []string{}
	for _, port := range appenderPorts {
		appender, err := appender.New(logger, storage, config.smallChunkMaxEntries, config.smallChunkSpreadMs, config.smallChunkMaxAgeMs)
		if err != nil {
			return nil, fmt.Errorf("unable to create appender %d: %v", port, err)
		}

		server, address, err := startAppenderServer(appender, port)
		if err != nil {
			return nil, fmt.Errorf("unable to start appender %d: %v", port, err)
		}
		servers = append(servers, server)
		appenders = append(appenders, appender)
		appenderAddresses = append(appenderAddresses, address)

		logger.Infof("Started appender at address: %s", address)
	}

	discovery, err := dc.New(appenderAddresses)
	if err != nil {
		return nil, fmt.Errorf("unable to create discovery: %v", err)
	}

	ingester, err := in.New(logger, discovery, appenderFanout)
	if err != nil {
		return nil, fmt.Errorf("unable to create ingester: %v", err)
	}

	return &localCluster{
		appenders: appenders,
		ingester:  ingester,
		storage:   storage,
		discovery: discovery,
		mixer:     mx.New(logger, storage, discovery),
		servers:   servers,
	}, nil
}

func (c *localCluster) stop() {
	for _, s := range c.servers {
		s.Stop()
	}
}

func startAppenderServer(appender *appender.Appender, port int) (*grpc.Server, string, error) {
	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen for port %d: %v", port, err)
	}

	server := grpc.NewServer()
	pb_almanac.RegisterAppenderServer(server, appender)
	go func() {
		server.Serve(listen)
	}()

	return server, listen.Addr().String(), nil
}

func newIngestRequest(e interface{}) (*pb_almanac.IngestRequest, error) {
	entryJson, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal entry to json: %v", err)
	}
	return &pb_almanac.IngestRequest{EntryJson: string(entryJson)}, nil
}
