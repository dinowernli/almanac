package cluster

import (
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

// Config holds a few configurable values defining the behavior of the system.
type Config struct {
	SmallChunkMaxEntries int
	SmallChunkSpreadMs   int64
	SmallChunkMaxAgeMs   int64

	StorageType string
	GcsBucket   string
}

// LocalCluster holds a test setup ready to use for testing.
type LocalCluster struct {
	Mixer    *mx.Mixer
	Ingester *in.Ingester

	Appenders []*appender.Appender
	Storage   *st.Storage
	Discovery *dc.Discovery

	servers []*grpc.Server
}

// CreateCluster sets up a test cluster, including all services required to run the system.
func CreateCluster(logger *logrus.Logger, config *Config, appenderPorts []int, appenderFanout int) (*LocalCluster, error) {
	var err error
	var storage *st.Storage
	if config.StorageType == storageTypeMemory {
		storage = st.NewInMemoryStorage()
	} else if config.StorageType == storageTypeGcs {
		storage, err = st.NewGcsStorage(config.GcsBucket)
		if err != nil {
			return nil, fmt.Errorf("unable to create gcs storage: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unrecognized storage type: %s", config.StorageType)
	}

	appenders := []*appender.Appender{}
	servers := []*grpc.Server{}
	appenderAddresses := []string{}
	for _, port := range appenderPorts {
		appender, err := appender.New(logger, storage, config.SmallChunkMaxEntries, config.SmallChunkSpreadMs, config.SmallChunkMaxAgeMs)
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

	return &LocalCluster{
		Appenders: appenders,
		Ingester:  ingester,
		Storage:   storage,
		Discovery: discovery,
		Mixer:     mx.New(logger, storage, discovery),

		servers: servers,
	}, nil
}

// Stop stops all the servers running as part of this local cluster.
func (c *LocalCluster) Stop() {
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