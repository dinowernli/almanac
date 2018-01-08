package appender

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
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
	maxSpreadMs   int64
	maxOpenTimeMs int64
}

// createOpenChunk creates a new openChunk instance containing the supplied
// log entry.
//
// - maxEntries is the maximum number of entries in this chunk before it gets closed.
// - maxSpreadMs is the maximum difference between the smallest and largest timestamp of entries in this chunk.
// - maxOpenTimeMs is a maximum duration for which the chunk will stay open.
// - sinkChannel is a channel the open chunk gets sent into once it is closed.
func createOpenChunk(entry *pb_almanac.LogEntry, maxEntries int, maxSpreadMs int64, maxOpenTimeMs int64, sinkChannel chan *openChunk) (*openChunk, error) {
	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}

	result := &openChunk{
		entries: map[string]*pb_almanac.LogEntry{},
		index:   index,
		chunkId: newChunkId(entry.TimestampMs),

		closed:      false,
		closeTimer:  nil,
		sinkChannel: sinkChannel,
		mutex:       &sync.Mutex{},

		maxEntries:    maxEntries,
		maxSpreadMs:   maxSpreadMs,
		maxOpenTimeMs: maxOpenTimeMs,
	}

	result.closeTimer = time.AfterFunc(time.Duration(maxOpenTimeMs)*time.Millisecond, result.close)

	return result, nil
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
	if newEndMs-newStartMs > c.maxSpreadMs {
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
	c.sinkChannel <- c
}

// newChunkId creates a new chunk id proto, starting out with the supplied
// timestamp.
func newChunkId(timestampMs int64) *pb_almanac.ChunkId {
	return &pb_almanac.ChunkId{
		Uid:     randomString(chunkUidLength),
		StartMs: timestampMs,
		EndMs:   timestampMs,
	}
}

// TODO(dino): Deduplicate these methods with appender.go.
// randomString produces a random string of lower case letters.
func randomString(num int) string {
	bytes := make([]byte, num)
	for i := 0; i < num; i++ {
		bytes[i] = byte(randomInt(97, 122)) // lowercase letters.
	}
	return string(bytes)
}

func randomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
