package main

import (
	"fmt"
	"net"
	"testing"

	"dinowernli.me/almanac/appender"
	dc "dinowernli.me/almanac/discovery"
	mx "dinowernli.me/almanac/mixer"
	pb_almanac "dinowernli.me/almanac/proto"
	st "dinowernli.me/almanac/storage"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestIntegration(t *testing.T) {
	entriesPerChunk := 10
	numAppenders := 5
	port := 51000

	storage := st.NewInMemoryStorage()

	appenders := []string{}
	for i := 0; i < numAppenders; i++ {
		appender, err := appender.New(fmt.Sprintf("appender-%d", i), storage, entriesPerChunk)
		assert.NoError(t, err)

		address := fmt.Sprintf("localhost:%d", port)
		port++

		err = startServer(appender, address)
		assert.NoError(t, err)
		appenders = append(appenders, address)
	}

	discovery := dc.New(appenders)
	mixer := mx.New(storage, discovery)

	mixer.Search(nil, nil)
}

func startServer(appender *appender.Appender, address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	server := grpc.NewServer()
	pb_logging.RegisterAppenderServer(server, appender)
	go func() {
		server.Serve(listen)
	}()

	return nil
}
