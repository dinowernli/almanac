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
	index bleve.Index
}

func (s *indexService) Search(ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
	log.Printf("handling search request: %v\n", request)
	bleveQuery := bleve.NewMatchQuery(request.Query)
	bleveSearch := bleve.NewSearchRequest(bleveQuery)
	bleveResult, err := s.index.Search(bleveSearch)
	if err != nil {
		log.Fatalf("failed to search: %v", err)
	}

	ids := []string{}
	for _, hit := range bleveResult.Hits {
		ids = append(ids, hit.ID)
	}

	return &pb_logging.SearchResponse{Ids: ids}, nil
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

	log.Println("starting grpc server")

	service := &indexService{index: index}
	server := grpc.NewServer()
	pb_logging.RegisterIndexServiceServer(server, service)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", 12345))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		server.Serve(listen)
	}()

	log.Println("dialing server")
	connection, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer connection.Close()

	remoteRequest := &pb_logging.SearchRequest{Query: "foo"}
	log.Println(remoteRequest)
	client := pb_logging.NewIndexServiceClient(connection)
	response, err := client.Search(context.Background(), remoteRequest)
	if err != nil {
		log.Fatalf("failed to make rpc: %v", err)
	}

	log.Printf("remote response: %v\n", response)
}
