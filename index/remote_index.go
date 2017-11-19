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

// remoteIndex is an implementation of Bleve's Index interface which delegates
// calls to a remote IndexService.
type remoteIndex struct {
	address string
}

func NewRemoteIndex(address string) *remoteIndex {
	return &remoteIndex{address: address}
}

func (i *remoteIndex) Search(req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	connection, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}
	defer connection.Close()

	bleveRequestBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marsal request: %v", err)
	}

	remoteRequest := &pb_logging.SearchRequest{BleveRequestBytes: bleveRequestBytes}
	client := pb_logging.NewIndexServiceClient(connection)
	response, err := client.Search(context.Background(), remoteRequest)
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

func (i *remoteIndex) SearchInContext(ctx context.Context, req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	return i.Search(req)
}

// Methods down here just make sure we implement the Index interface.

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
