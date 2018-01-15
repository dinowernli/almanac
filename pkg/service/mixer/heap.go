package mixer

import (
	"fmt"

	st "dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"golang.org/x/net/context"
)

// heapItem represent an entry in the heap during merging.
type heapItem interface {
	// key returns the key which should be used to sort this item.
	key() int64

	// entry returns the current log entry associated with this item.
	entry() (*pb_almanac.LogEntry, error)

	// next returns a heapItem which goes back into the heap once this
	// item has been dealt with. Returns nil if there is no next item.
	next() (heapItem, error)
}

// chunkHeapItem is a HeapItem backed by a chunk in storage.
type chunkHeapItem struct {
	chunkIdProto *pb_almanac.ChunkId
	searchRequest *pb_almanac.SearchRequest
	ctx context.Context
	storage *st.Storage

	entries []*pb_almanac.LogEntry
	idx int
	loaded bool
}

func (i *chunkHeapItem) key() int64 {
	// Determine the first key from the chunk id, keeping the first heap item cheap.
	if i.idx == 0 {
		return i.chunkIdProto.StartMs
	}

	// If this is not the first item, we've loaded the entries.
	return i.entries[i.idx].TimestampMs
}

func (i *chunkHeapItem) entry() (*pb_almanac.LogEntry, error) {
	err := i.ensureChunkLoaded()
	if err != nil {
		return nil, fmt.Errorf("unable to load chunk from 'entry': %v", err)
	}
	return i.entries[i.idx], nil
}

func (i *chunkHeapItem) next() (heapItem, error) {
	err := i.ensureChunkLoaded()
	if err != nil {
		return nil, fmt.Errorf("unable to load chunk from 'next' : %v", err)
	}

	if i.idx >= len(i.entries) {
		// We've used up all the results, no next item.
		return nil, nil
	}

	// Just update the index and reuse this item as the next one.
	i.idx++
	return i, nil
}

// ensureChunkLoaded makes a request to storage to load the entries associated with this chunk
// if the entries haven't already been loaded. If no error is returned, then "loaded is true
// and "entries" is populated after calling this.
func (i *chunkHeapItem) ensureChunkLoaded() error {
	if i.loaded {
		return nil
	}

	chunkId, err := st.ChunkId(i.chunkIdProto)
	if err != nil {
		return fmt.Errorf("unable to compute chunk id from proto: %v", err)
	}

	chunk, err := i.storage.LoadChunk(chunkId)
	if err != nil {
		return fmt.Errorf("unable to load chunk %s from storage: %v", chunkId, err)
	}

	entries, err := chunk.Search(i.ctx, i.searchRequest.Query, i.searchRequest.Num, i.searchRequest.StartMs, i.searchRequest.EndMs)
	if err != nil {
		return fmt.Errorf("unable to search chunk: %v", err)
	}

	i.entries = entries
	i.loaded = true
	return nil
}

// appenderHeapItem is a HeapItem backed by an appender grpc service.
type appenderHeapItem struct {
	entries []*pb_almanac.LogEntry
	idx int
}

func (i *appenderHeapItem) key() int64 {
	return i.entries[i.idx].TimestampMs
}

func (i *appenderHeapItem) entry() (*pb_almanac.LogEntry, error) {
	return i.entries[i.idx], nil
}

func (i *appenderHeapItem) next() (heapItem, error) {
	if i.idx >= len(i.entries) {
		// We've used up all the results, no next item.
		return nil, nil
	}

	// Just update the index and reuse this item as the next one.
	i.idx++
	return i, nil
}