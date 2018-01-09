package main

import (
	"encoding/json"
	"fmt"
	"net"

	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/service/appender"
	dc "dinowernli.me/almanac/service/discovery"
	in "dinowernli.me/almanac/service/ingester"
	mx "dinowernli.me/almanac/service/mixer"
	st "dinowernli.me/almanac/storage"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
func createCluster(logger *logrus.Logger, config *config, appenderAddresses []string, appenderFanout int) (*localCluster, error) {
	storage := st.NewInMemoryStorage()
	appenders := []*appender.Appender{}
	servers := []*grpc.Server{}
	for _, address := range appenderAddresses {
		appender, err := appender.New(logger, storage, config.smallChunkMaxEntries, config.smallChunkSpreadMs, config.smallChunkMaxAgeMs)
		if err != nil {
			return nil, fmt.Errorf("unable to create appender %s: %v", address, err)
		}

		server, err := startAppenderServer(appender, address)
		if err != nil {
			return nil, fmt.Errorf("unable to start appender %s: %v", address, err)
		}
		servers = append(servers, server)
		appenders = append(appenders, appender)

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

func startAppenderServer(appender *appender.Appender, address string) (*grpc.Server, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for address %s: %v", address, err)
	}

	server := grpc.NewServer()
	pb_almanac.RegisterAppenderServer(server, appender)
	go func() {
		server.Serve(listen)
	}()

	return server, nil
}

func newIngestRequest(e interface{}) (*pb_almanac.IngestRequest, error) {
	entryJson, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal entry to json: %v", err)
	}
	return &pb_almanac.IngestRequest{EntryJson: string(entryJson)}, nil
}
