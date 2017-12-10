package index

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	pb_logging "dinowernli.me/almanac/proto"

	"github.com/blevesearch/bleve"
	"golang.org/x/net/context"
)

// Index implements a grpc service representing a remote index.
type Index struct {
	index bleve.Index
	path  string
}

func openIndexService(dir string) (*Index, error) {
	index, err := bleve.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &Index{index: index, path: dir}, nil
}

func NewIndexService() (*Index, error) {
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
	return &Index{index: index, path: dir}, nil
}

func (s *Index) Index(id string, data interface{}) error {
	return s.index.Index(id, data)
}

func (s *Index) Search(ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
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
