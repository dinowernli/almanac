package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	pb_logging "dinowernli.me/logging-tmp/proto"

	"github.com/blevesearch/bleve"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type data struct {
	Name string
}

// remoteIndex implements Bleve's Index interface, backed by a remote
// implementation of the index service.
type remoteIndex struct {
}

// indexService implements a grpc service representing a remote index.
type indexService struct {
}

func (s *indexService) Search(ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
	log.Println("handling search request: %v", request)
	return &pb_logging.SearchResponse{}, nil
}

func main() {
	dir, err := ioutil.TempDir("", "index.bleve")
	if err != nil {
		log.Fatalf("failed to create tempfile: %v", err)
	}
	log.Printf("created directory: %s", dir)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dir, mapping)
	if err != nil {
		log.Fatalf("failed to create index: %v", err)
	}

	index.Index("id1", &data{Name: "foo"})
	index.Index("id2", &data{Name: "bar"})

	query := bleve.NewMatchQuery("foo")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		log.Fatalf("failed to search: %v", err)
	}

	log.Println(searchResults)

	remoteRequest := &pb_logging.SearchRequest{
		Query: "foo",
	}
	log.Println(remoteRequest)

	log.Println("starting grpc server")

	server := grpc.NewServer()
	pb_logging.RegisterIndexServiceServer(server, &indexService{})
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", 12345))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	server.Serve(listen)
}
