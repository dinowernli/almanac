package storage

import (
	"fmt"
	"strings"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"

	"golang.org/x/net/context"
)

const (
	chunkIdFormat    = "%d-%d-%s"
	chunkIdSeparator = "-"
)

// ChunkId returns the string representation of the supplied chunk id proto.
func ChunkId(idProto *pb_almanac.ChunkId) (string, error) {
	if idProto.Uid == "" {
		return "", fmt.Errorf("cannot create chunk id with empty uid")
	}

	if strings.Contains(idProto.Uid, chunkIdSeparator) {
		return "", fmt.Errorf("chunk uid cannot contain '-', but got: %s", idProto.Uid)
	}

	if idProto.StartMs > idProto.EndMs {
		return "", fmt.Errorf("invalid start and end times: start=%d, end=%d", idProto.StartMs, idProto.EndMs)
	}

	return fmt.Sprintf(chunkIdFormat, idProto.StartMs, idProto.EndMs, idProto.Uid), nil
}

// ChunkIdProto returns the structured representation of the supplied chunk id.
func ChunkIdProto(chunkId string) (*pb_almanac.ChunkId, error) {
	var uid string
	var startMs int64
	var endMs int64
	_, err := fmt.Sscanf(chunkId, chunkIdFormat, &startMs, &endMs, &uid)
	if err != nil {
		return nil, fmt.Errorf("unable to parse id %s: %v", chunkId, err)
	}

	return &pb_almanac.ChunkId{StartMs: startMs, EndMs: endMs, Uid: uid}, nil
}

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
		return nil, fmt.Errorf("failed to deserialize index from chunk %s: %v", chunkId, err)
	}

	entryMap := map[string]*pb_almanac.LogEntry{}
	for _, entry := range chunkProto.Entries {
		entryMap[entry.Id] = entry
	}
	return &Chunk{id: chunkId, index: idx, entries: entryMap}, nil
}

// Search returns all log entries in the chunk matching the supplied query.
func (c *Chunk) Search(ctx context.Context, query string, num int32, startMs int64, endMs int64) ([]*pb_almanac.LogEntry, error) {
	return Search(ctx, c.index, c.entries, query, num, startMs, endMs)
}

// Close releases any resources associated with this chunk.
func (c *Chunk) Close() error {
	return c.index.Close()
}

// Search executes a search on a given index and set of entries, returning results
// in arbitrary order.
func Search(ctx context.Context, idx *index.Index, entries map[string]*pb_almanac.LogEntry, query string, num int32, startMs int64, endMs int64) ([]*pb_almanac.LogEntry, error) {
	ids, err := idx.Search(ctx, query, num)
	if err != nil {
		return nil, fmt.Errorf("unable to search index: %v", err)
	}

	result := []*pb_almanac.LogEntry{}
	for _, id := range ids {
		entry, ok := entries[id]
		if !ok {
			return nil, fmt.Errorf("could not locate hit %s", id)
		}

		if startMs != 0 && entry.TimestampMs < startMs {
			continue
		}
		if endMs != 0 && entry.TimestampMs > endMs {
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}

func (c *Chunk) fetchEntry(id string) (*pb_almanac.LogEntry, error) {
	result, ok := c.entries[id]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", id)
	}
	return result, nil
}
