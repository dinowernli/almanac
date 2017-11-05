package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	pb_logging "dinowernli.me/logging-tmp/proto"

	"github.com/blevesearch/bleve"
	"golang.org/x/net/context"
)

// indexService implements a grpc service representing a remote index.
type indexService struct {
	index bleve.Index
}

func (s *indexService) Index(id string, data interface{}) error {
	return s.index.Index(id, data)
}

func newIndexService() (*indexService, error) {
	dir, err := ioutil.TempDir("", "index.bleve")
	if err != nil {
		return nil, fmt.Errorf("failed to create tempfile: %v", err)
	}
	log.Printf("created directory: %s", dir)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dir, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &indexService{index: index}, nil
}

func (s *indexService) Search(ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
	bleveSearchRequest := &bleve.SearchRequest{}
	err := json.Unmarshal(request.BleveRequestBytes, bleveSearchRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal request: %v", err)
	}

	bleveResult, err := s.index.Search(bleveSearchRequest)
	if err != nil {
		log.Fatalf("failed to search: %v", err)
	}

	bleveResultBytes, err := json.Marshal(bleveResult)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal response: %v", err)
	}
	return &pb_logging.SearchResponse{BleveResponseBytes: bleveResultBytes}, nil
}
