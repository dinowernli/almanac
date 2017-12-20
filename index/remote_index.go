package index

import (
	"encoding/json"
	"fmt"

	pb_logging "dinowernli.me/almanac/proto"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/mapping"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// searchClient represents the interface needed to perform remote searches.
type searchClient interface {
	Search(ctx context.Context, request *pb_logging.SearchRequest, opts ...grpc.CallOption) (*pb_logging.SearchResponse, error)
}

// remoteIndex is an implementation of Bleve's Index interface which delegates
// calls to a remote IndexService.
type remoteIndex struct {
	client searchClient
}

func NewRemoteIndex(client searchClient) *Index {
	// TODO(dino): This is not great, remove the empty path stuff once Index becomes an interface.
	return &Index{index: &remoteIndex{client: client}, path: ""}
}

func (i *remoteIndex) SearchInContext(ctx context.Context, req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	bleveRequestBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marsal request: %v", err)
	}

	remoteRequest := &pb_logging.SearchRequest{BleveRequestBytes: bleveRequestBytes}
	response, err := i.client.Search(ctx, remoteRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to make rpc: %v", err)
	}

	result := &bleve.SearchResult{}
	err = json.Unmarshal(response.BleveResponseBytes, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %v", err)
	}
	return result, nil
}

// Methods down here just make sure we implement the Index interface.

func (i *remoteIndex) Search(req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	return nil, fmt.Errorf("Search without context not implemented")
}

func (i *remoteIndex) Advanced() (index.Index, store.KVStore, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) Mapping() mapping.IndexMapping {
	return nil
}

func (i *remoteIndex) Index(id string, data interface{}) (err error) {
	return fmt.Errorf("not implemented")
}

func (i *remoteIndex) Delete(id string) (err error) {
	return fmt.Errorf("not implemented")
}

func (i *remoteIndex) Batch(b *bleve.Batch) error {
	return fmt.Errorf("not implemented")
}

func (i *remoteIndex) Document(id string) (doc *document.Document, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) DocCount() (uint64, error) {
	return 0, fmt.Errorf("DocCount not implemented")
}
func (i *remoteIndex) Fields() (fields []string, err error) {
	return nil, fmt.Errorf("Fields not implemented")
}

func (i *remoteIndex) FieldDict(field string) (index.FieldDict, error) {
	return nil, fmt.Errorf("FieldDict not implemented")
}

func (i *remoteIndex) FieldDictRange(field string, startTerm []byte, endTerm []byte) (index.FieldDict, error) {
	return nil, fmt.Errorf("FieldDictRange not implemented")
}

func (i *remoteIndex) FieldDictPrefix(field string, termPrefix []byte) (index.FieldDict, error) {
	return nil, fmt.Errorf("FieldDictPrefix not implemented")
}

func (i *remoteIndex) Close() error {
	return fmt.Errorf("Close not implemented")
}

func (i *remoteIndex) Stats() *bleve.IndexStat {
	return nil
}

func (i *remoteIndex) StatsMap() map[string]interface{} {
	return nil
}

func (i *remoteIndex) GetInternal(key []byte) (val []byte, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) SetInternal(key, val []byte) error {
	return fmt.Errorf("not implemented")
}

func (i *remoteIndex) DeleteInternal(key []byte) error {
	return fmt.Errorf("not implemented")
}

func (i *remoteIndex) NewBatch() *bleve.Batch {
	return nil
}

func (i *remoteIndex) Name() string {
	return "some-id"
}

func (i *remoteIndex) SetName(name string) {
}
