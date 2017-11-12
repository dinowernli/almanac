package main

import (
	"fmt"

	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/registry"
)

const (
	StoreName = "store"
)

type myStore struct {
}

func NewStore(mergeOperator store.MergeOperator, config map[string]interface{}) (store.KVStore, error) {
	return nil, fmt.Errorf("store not implemented")
}

func RegisterStore() {
	registry.RegisterKVStore(StoreName, NewStore)
}
