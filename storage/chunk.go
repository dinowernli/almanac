package storage

import (
	"fmt"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/blevesearch/bleve"
)

// Chunk is an in-memory, immutable representation of a stored chunk. A chunk
// must be closed by calling Close() once it is no longer in use.
type Chunk struct {
	id      string
	index   *index.Index
	entries map[string]*pb_almanac.LogEntry

	// TODO(dino): Use a field to detect use-after-close bugs.
}

// openChunk returns a chunk instance for the supplied proto. The caller is
// responsible for calling Close() on the returned instance.
func openChunk(chunkId string, chunkProto *pb_almanac.Chunk) (*Chunk, error) {
	idx, err := index.Deserialize(chunkProto.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize index from chunk: %v", chunkId, err)
	}

	entryMap := map[string]*pb_almanac.LogEntry{}
	for _, entry := range chunkProto.Entries {
		entryMap[entry.Id] = entry
	}
	return &Chunk{id: chunkId, index: idx, entries: entryMap}, nil
}

// Search returns all log entries in the chunk matching the supplied query.
func (c *Chunk) Search(query string, num int32) ([]*pb_almanac.LogEntry, error) {
	request := bleve.NewSearchRequestOptions(
		bleve.NewMatchQuery(query),
		int(num),
		0, /* from */
		false /* explain */)

	result, err := c.index.Bleve().Search(request)
	if err != nil {
		return nil, fmt.Errorf("unable to search index: %v", err)
	}

	entries := []*pb_almanac.LogEntry{}
	for _, hit := range result.Hits {
		entry, ok := c.entries[hit.ID]
		if !ok {
			return nil, fmt.Errorf("could not locate hit %s", hit.ID)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Close releases any resources associated with this chunk.
func (c *Chunk) Close() error {
	return c.index.Close()
}

func (c *Chunk) fetchEntry(id string) (*pb_almanac.LogEntry, error) {
	result, ok := c.entries[id]
	if !ok {
		return nil, fmt.Errorf("entry %d not found", id)
	}
	return result, nil
}
