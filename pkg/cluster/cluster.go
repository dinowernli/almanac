package cluster

import (
	"fmt"
	"net"
	"time"

	"github.com/dinowernli/almanac/pkg/service/appender"
	dc "github.com/dinowernli/almanac/pkg/service/discovery"
	in "github.com/dinowernli/almanac/pkg/service/ingester"
	"github.com/dinowernli/almanac/pkg/service/janitor"
	mx "github.com/dinowernli/almanac/pkg/service/mixer"
	st "github.com/dinowernli/almanac/pkg/storage"
	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Config holds a few configurable values defining the behavior of the system.
type Config struct {
	SmallChunkMaxEntries int
	SmallChunkSpread     time.Duration
	SmallChunkMaxAge     time.Duration

	BigChunkMaxSpread time.Duration

	JanitorCompactionInterval time.Duration

	StorageType string
	GcsBucket   string
	DiskPath    string
}

// LocalCluster holds a test setup ready to use for testing.
type LocalCluster struct {
	Mixer    *mx.Mixer
	Ingester *in.Ingester

	Janitor   *janitor.Janitor
	Appenders []*appender.Appender
	Storage   *st.Storage
	Discovery *dc.Discovery

	servers []*grpc.Server
}

// CreateCluster sets up a test cluster, including all services required to run the system.
func CreateCluster(ctx context.Context, logger *logrus.Logger, config *Config, appenderPorts []int, ingestFanout int) (*LocalCluster, error) {
	var err error
	var storage *st.Storage
	if config.StorageType == st.StorageTypeMemory {
		storage, err = st.NewMemoryStorage()
		if err != nil {
			return nil, fmt.Errorf("unable to create memory storage: %v", err)
		}
	} else if config.StorageType == st.StorageTypeGcs {
		storage, err = st.NewGcsStorage(config.GcsBucket)
		if err != nil {
			return nil, fmt.Errorf("unable to create gcs storage: %v", err)
		}
	} else if config.StorageType == st.StorageTypeDisk {
		storage, err = st.NewDiskStorage(config.DiskPath)
		if err != nil {
			return nil, fmt.Errorf("unable to create disk storage: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unrecognized storage type: %s", config.StorageType)
	}

	appenders := []*appender.Appender{}
	servers := []*grpc.Server{}
	appenderAddresses := []string{}
	for _, port := range appenderPorts {
		appender, err := appender.New(logger, storage, config.SmallChunkMaxEntries, config.SmallChunkSpread, config.SmallChunkMaxAge)
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

	ingester, err := in.New(logger, discovery, ingestFanout)
	if err != nil {
		return nil, fmt.Errorf("unable to create ingester: %v", err)
	}

	janitor, err := janitor.New(ctx, logger, storage, config.JanitorCompactionInterval, config.BigChunkMaxSpread)
	if err != nil {
		return nil, fmt.Errorf("unable to create janitor: %v", err)
	}

	return &LocalCluster{
		Appenders: appenders,
		Ingester:  ingester,
		Janitor:   janitor,
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
