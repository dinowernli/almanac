package main

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/mapping"
	"golang.org/x/net/context"
)

// remoteIndex is an implementation of Bleve's Index interface which delegates
// calls to a remote IndexService.
type remoteIndex struct {
	address string
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
	return 0, fmt.Errorf("not implemented")
}

func (i *remoteIndex) Search(req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) SearchInContext(ctx context.Context, req *bleve.SearchRequest) (sr *bleve.SearchResult, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) Fields() (fields []string, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) FieldDict(field string) (index.FieldDict, error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) FieldDictRange(field string, startTerm []byte, endTerm []byte) (index.FieldDict, error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) FieldDictPrefix(field string, termPrefix []byte) (index.FieldDict, error) {
	return nil, fmt.Errorf("not implemented")
}

func (i *remoteIndex) Close() error {
	return fmt.Errorf("not implemented")
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
