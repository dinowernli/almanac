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

	"google.golang.org/grpc"
)

// entry represents a log entry as would be supplied by a user of the system.
type entry struct {
	Message     string `json:"message"`
	Logger      string `json:"logger"`
	TimestampMs int64  `json:"timestamp_ms"`
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
func createCluster(startPort int, numAppenders int, entriesPerChunk int, appenderFanout int) (*localCluster, error) {
	nextPort := startPort

	storage := st.NewInMemoryStorage()
	appenderAddresses := []string{}
	appenders := []*appender.Appender{}
	servers := []*grpc.Server{}
	for i := 0; i < numAppenders; i++ {
		appender, err := appender.New(storage, entriesPerChunk)
		if err != nil {
			return nil, fmt.Errorf("unable to create appender %d: %v", i, err)
		}

		address := fmt.Sprintf("localhost:%d", nextPort)
		nextPort++

		server, err := startAppenderServer(appender, address)
		if err != nil {
			return nil, fmt.Errorf("unable to start appender %d: %v", i, err)
		}
		servers = append(servers, server)

		appenderAddresses = append(appenderAddresses, address)
		appenders = append(appenders, appender)
	}

	discovery, err := dc.New(appenderAddresses)
	if err != nil {
		return nil, fmt.Errorf("unable to create discovery: %v", err)
	}

	ingester, err := in.New(discovery, appenderFanout)
	if err != nil {
		return nil, fmt.Errorf("unable to create ingester: %v", err)
	}

	return &localCluster{
		appenders: appenders,
		ingester:  ingester,
		storage:   storage,
		discovery: discovery,
		mixer:     mx.New(storage, discovery),
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

func ingestRequest(e interface{}) (*pb_almanac.IngestRequest, error) {
	entryJson, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal entry to json: %v", err)
	}
	return &pb_almanac.IngestRequest{EntryJson: string(entryJson)}, nil
}
