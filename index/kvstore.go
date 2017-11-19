package index

import (
	"fmt"

	pb_logging "dinowernli.me/almanac/proto"

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

// SerializeStore returns a proto contains all the (key, value) pairs present
// in the supplied store.
func SerializeStore(kvstore store.KVStore) (*pb_logging.Store, error) {
	reader, err := kvstore.Reader()
	if err != nil {
		return nil, fmt.Errorf("unable to get reader: %v", err)
	}

	entries := []*pb_logging.Entry{}
	iterator := reader.PrefixIterator([]byte{})
	for iterator.Valid() {
		entries = append(entries, &pb_logging.Entry{
			Key:   iterator.Key(),
			Value: iterator.Value(),
		})
	}

	return &pb_logging.Store{Entries: entries}, nil
}

// DeserializeStore add all the contents of the supplied store proto to the
// supplied kvstore.
func DeserializeStore(proto *pb_logging.Store, kvstore store.KVStore) error {
	writer, err := kvstore.Writer()
	if err != nil {
		return fmt.Errorf("unable to get writer: %v", err)
	}

	batch := writer.NewBatch()
	for _, entry := range proto.Entries {
		batch.Set(entry.Key, entry.Value)
	}

	err = writer.ExecuteBatch(batch)
	if err != nil {
		return fmt.Errorf("unable to execute set batch: %v", err)
	}
	return nil
}
