package main

import (
	"fmt"
	"log"
	"net"

	"dinowernli.me/almanac/appender"
	"dinowernli.me/almanac/index"
	pb_logging "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/blevesearch/bleve"
	"google.golang.org/grpc"
)

const (
	maxEntriesPerChunk = 4
)

type data struct {
	Name string
}

func main() {
	diskStorage, err := storage.NewTempDiskStorage()
	if err != nil {
		log.Fatalf("unable to create storage: %v", err)
	}

	err = diskStorage.Write("fileid", []byte("hello"))
	if err != nil {
		log.Fatalf("unable to write to storage: %v", err)
	}

	service, err := index.NewIndex()
	if err != nil {
		log.Fatalf("failed to create index service: %v", err)
	}
	service.Index("id1", &data{Name: "foo"})
	service.Index("id2", &data{Name: "bar"})
	log.Println("created index")

	server := grpc.NewServer()
	pb_logging.RegisterIndexServiceServer(server, service)

	appender, err := appender.New(diskStorage, maxEntriesPerChunk)
	if err != nil {
		log.Fatalf("failed to create appender: %v", err)
	}

	pb_logging.RegisterAppenderServer(server, appender)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", 12345))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		server.Serve(listen)
	}()
	log.Println("started grpc server")

	indexAlias := bleve.NewIndexAlias(index.NewRemoteIndex("localhost:12345"))
	bleveQuery := bleve.NewMatchQuery("foo")
	bleveSearch := bleve.NewSearchRequest(bleveQuery)
	bleveResult, err := indexAlias.Search(bleveSearch)
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	log.Printf("done searching remotely: %v", bleveResult)
}
