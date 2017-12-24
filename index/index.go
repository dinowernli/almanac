package index

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb_logging "dinowernli.me/almanac/proto"

	"github.com/blevesearch/bleve"
	"golang.org/x/net/context"
)

// Index wraps a bleve index and presents an interface in terms of the almanac
// protos. Instances of Index can be serialized to protos and deserialized from
// protos.
type Index struct {
	index bleve.Index
	path  string
}

// openIndex returns an index backed by the contents on disk at the specified
// path. The caller responsible for eventually calling Close() on the returned
// instance to release disk resources.
func openIndex(dir string) (*Index, error) {
	index, err := bleve.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &Index{index: index, path: dir}, nil
}

// NewIndex returns an instance of index backed by an temporary location on disk.
// The caller is responsible for eventually calling Close() on the returned index.
func NewIndex() (*Index, error) {
	dir, err := ioutil.TempDir("", "index.bleve")
	if err != nil {
		return nil, fmt.Errorf("failed to create tempfile: %v", err)
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dir, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &Index{index: index, path: dir}, nil
}

func (i *Index) Bleve() bleve.Index {
	return i.index
}

func (i *Index) Index(id string, data interface{}) error {
	return i.index.Index(id, data)
}

func (i *Index) Search(ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
	return Search([]*Index{i}, ctx, request)
}

// Close releases any resources held by this instance. No other methods must
// be called after this.
func (i *Index) Close() error {
	return os.RemoveAll(i.path)
}

// Search executes the supplied search request on the supplied bleve index.
func Search(indexes []*Index, ctx context.Context, request *pb_logging.SearchRequest) (*pb_logging.SearchResponse, error) {
	bleveSearchRequest := &bleve.SearchRequest{}
	err := json.Unmarshal(request.BleveRequestBytes, bleveSearchRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal request: %v", err)
	}

	bleveIndexes := []bleve.Index{}
	for _, idx := range indexes {
		bleveIndexes = append(bleveIndexes, idx.index)
	}
	indexAlias := bleve.NewIndexAlias(bleveIndexes...)

	bleveResult, err := indexAlias.Search(bleveSearchRequest)
	if err != nil {
		log.Fatalf("failed to search: %v", err)
	}

	bleveResultBytes, err := json.Marshal(bleveResult)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal response: %v", err)
	}
	return &pb_logging.SearchResponse{BleveResponseBytes: bleveResultBytes}, nil
}
