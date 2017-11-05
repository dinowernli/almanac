package main

import (
	"fmt"
	"log"
	"net"

	pb_logging "dinowernli.me/logging-tmp/proto"

	"github.com/blevesearch/bleve"
	"google.golang.org/grpc"
)

type data struct {
	Name string
}

func main() {
	service, err := newIndexService()
	if err != nil {
		log.Fatalf("failed to create index service: %v", err)
	}
	service.Index("id1", &data{Name: "foo"})
	service.Index("id2", &data{Name: "bar"})
	log.Println("created index")

	server := grpc.NewServer()
	pb_logging.RegisterIndexServiceServer(server, service)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", 12345))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		server.Serve(listen)
	}()
	log.Println("started grpc server")

	indexAlias := bleve.NewIndexAlias(&remoteIndex{address: "localhost:12345"})
	bleveQuery := bleve.NewMatchQuery("foo")
	bleveSearch := bleve.NewSearchRequest(bleveQuery)
	bleveResult, err := indexAlias.Search(bleveSearch)
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	log.Printf("done searching remotely: %v", bleveResult)
}
