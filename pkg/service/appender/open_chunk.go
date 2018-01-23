package appender

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"dinowernli.me/almanac/pkg/index"
	"dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"golang.org/x/net/context"
)

// openChunk holds the data for a chunk under construction.
type openChunk struct {
	entries map[string]*pb_almanac.LogEntry
	index   *index.Index
	chunkId *pb_almanac.ChunkId

	closed      bool
	closeTimer  *time.Timer
	sinkChannel chan *openChunk
	mutex       *sync.Mutex

	maxEntries    int
	maxSpread     time.Duration
	maxOpenTimeMs int64
}

// newOpenChunk creates a new openChunk instance containing the supplied log entry.
//
// - maxEntries is the maximum number of entries in this chunk before it gets closed.
// - maxSpread is the maximum difference between the smallest and largest timestamp of entries in this chunk.
// - maxOpenTimeMs is a maximum duration for which the chunk will stay open.
// - sinkChannel is a channel the open chunk gets sent into once it is closed.
func newOpenChunk(entry *pb_almanac.LogEntry, maxEntries int, maxSpread time.Duration, maxOpenTimeMs int64, sinkChannel chan *openChunk) (*openChunk, error) {
	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}
	result := &openChunk{
		entries: map[string]*pb_almanac.LogEntry{},
		index:   index,
		chunkId: newChunkId(),

		closed:      false,
		closeTimer:  nil,
		sinkChannel: sinkChannel,
		mutex:       &sync.Mutex{},

		maxEntries:    maxEntries,
		maxSpread:     maxSpread,
		maxOpenTimeMs: maxOpenTimeMs,
	}

	// Add the first entry.
	added, err := result.tryAdd(entry)
	if err != nil {
		return nil, fmt.Errorf("unable to add first entry to chunk: %v", err)
	}
	if !added {
		// Indicates a programming error. The first addition should always go through.
		return nil, fmt.Errorf("empty chunk rejected entry")
	}

	// Make sure we respect the maximum lifetime of the open chunk.
	result.closeTimer = time.AfterFunc(time.Duration(maxOpenTimeMs)*time.Millisecond, result.close)

	return result, nil
}

// search executes a search on the in-memory entries and return the matching results
// (in arbitrary order).
func (c *openChunk) search(ctx context.Context, request *pb_almanac.SearchRequest) ([]*pb_almanac.LogEntry, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return storage.Search(ctx, c.index, c.entries, request.Query, request.Num, request.StartMs, request.EndMs)
}

// tryAdd attempts to add the supplied entry to the chunk.
//
// Return values are to be interpreted as follows:
// - the first return value indicates whether the entry was added.
// - the second return value indicates whether there was an error while trying to append.
func (c *openChunk) tryAdd(entry *pb_almanac.LogEntry) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we are already closed.
	if c.closed {
		return false, nil
	}

	// Check if we would be violating the spread contraints.
	newStartMs := c.chunkId.StartMs
	if entry.TimestampMs < newStartMs {
		newStartMs = entry.TimestampMs
	}
	newEndMs := c.chunkId.EndMs
	if entry.TimestampMs > newEndMs {
		newEndMs = entry.TimestampMs
	}
	if time.Duration(newEndMs-newStartMs)*time.Millisecond > c.maxSpread {
		return false, nil
	}

	// Parse and index the entry.
	var rawEntry interface{}
	err := json.Unmarshal([]byte(entry.EntryJson), &rawEntry)
	if err != nil {
		return false, fmt.Errorf("unable to parse raw json: %v", err)
	}

	err = c.index.Index(entry.Id, rawEntry)
	if err != nil {
		return false, fmt.Errorf("unable to index raw json entry: %v", err)
	}
	c.entries[entry.Id] = entry

	// Update the timestamps.
	c.chunkId.StartMs = newStartMs
	c.chunkId.EndMs = newEndMs

	// Check if this chunk needs to get closed.
	if len(c.entries) >= c.maxEntries {
		go c.close()
	}

	return true, nil
}

// close marks this chunk as closed and sends it through the sink channel for
// processing.
func (c *openChunk) close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// There are multiple competing reasons to close a chunk, so make sure it
	// only happens once.
	if c.closed {
		return
	}

	c.closeTimer.Stop()
	c.closed = true

	go func() {
		c.sinkChannel <- c
	}()
}

// toProto turns this instance into a chunk proto. This must only be called
// after closing this openChunk.
func (c *openChunk) toProto() (*pb_almanac.Chunk, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		return nil, fmt.Errorf("must only call toProto() after closing")
	}

	indexProto, err := index.Serialize(c.index)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize index: %v", err)
	}
	c.index.Close()

	entries := []*pb_almanac.LogEntry{}
	for _, e := range c.entries {
		entries = append(entries, e)
	}

	return &pb_almanac.Chunk{
		Id:      c.chunkId,
		Entries: entries,
		Index:   indexProto,
	}, nil
}

// newChunkId creates a new chunk id proto, starting out with the supplied
// timestamp.
func newChunkId() *pb_almanac.ChunkId {
	return &pb_almanac.ChunkId{
		Uid:     storage.NewChunkUid(),
		StartMs: math.MaxInt64,
		EndMs:   math.MinInt64,
	}
}
