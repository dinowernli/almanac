package appender

import (
	"encoding/json"
	"fmt"
	"sync"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// appender keeps track of a single open chunk under construction at a time and
// periodically decides that the open chunk is complete, at which point the
// appender writes the chunk to storage. The appender also knows how to answer
// search requests for the currently open chunk.
type Appender struct {
	entries     map[string]*pb_almanac.LogEntry
	index       *index.Index
	chunkId     int
	appendMutex *sync.Mutex

	appenderId         string
	storage            *storage.Storage
	maxEntriesPerChunk int
}

// New returns a new appender backed by the supplied storage.
func New(appenderId string, storage *storage.Storage, maxEntriesPerChunk int) (*Appender, error) {
	if maxEntriesPerChunk < 1 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxEntriesPerChunk)
	}

	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}

	return &Appender{
		entries:            map[string]*pb_almanac.LogEntry{},
		index:              index,
		appendMutex:        &sync.Mutex{},
		storage:            storage,
		maxEntriesPerChunk: maxEntriesPerChunk,
		appenderId:         appenderId,
	}, nil
}

func (a *Appender) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	ids, err := a.index.Search(ctx, request.Query, request.Num)
	if err != nil {
		return nil, fmt.Errorf("unable to search index: %v", err)
	}

	entries := []*pb_almanac.LogEntry{}
	for _, id := range ids {
		entry, ok := a.entries[id]
		if !ok {
			return nil, fmt.Errorf("could not locate hit %s", id)
		}
		entries = append(entries, entry)
	}

	return &pb_almanac.SearchResponse{Entries: entries}, nil
}

func (a *Appender) Append(ctx context.Context, request *pb_almanac.AppendRequest) (*pb_almanac.AppendResponse, error) {
	// Perform some validation of the request.
	logEntry := request.GetEntry()
	if logEntry == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "no entry supplied")
	}

	entryId := logEntry.GetId()
	if entryId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "no id supplied")
	}

	// Update the data structures corresponding to the open chunk.
	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	var rawEntry interface{}
	err := json.Unmarshal([]byte(logEntry.EntryJson), &rawEntry)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "unable to parse raw json: %v", err)
	}

	err = a.index.Index(entryId, rawEntry)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to index raw json entry: %v", err)
	}

	a.entries[entryId] = request.Entry

	// If the open chunk has grown enough, close it up and start a new one.
	if len(a.entries) >= a.maxEntriesPerChunk {
		err := a.storeChunk()
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "unable to store chunk: %v", err)
		}

		index, err := index.NewIndex()
		if err != nil {
			// TODO(dino): If this happens, stop responding to requests.
			return nil, grpc.Errorf(codes.Internal, "unable to create new index: %v", err)
		}
		a.index = index
		a.entries = map[string]*pb_almanac.LogEntry{}
	}

	return &pb_almanac.AppendResponse{}, nil
}

// storeChunk takes the currently open chunk, persists it to storage, and starts
// a new open chunk.
func (a *Appender) storeChunk() error {
	indexProto, err := index.Serialize(a.index)
	if err != nil {
		return fmt.Errorf("unable to serialize index: %v", err)
	}

	entries := []*pb_almanac.LogEntry{}
	for _, e := range a.entries {
		entries = append(entries, e)
	}

	chunkProto := &pb_almanac.Chunk{
		Entries: entries,
		Index:   indexProto,
	}

	chunkName := a.nextChunkName()
	err = a.storage.StoreChunk(chunkName, chunkProto)
	if err != nil {
		return fmt.Errorf("unable to store chunk %s: %v", chunkName, err)
	}
	return nil
}

func (a *Appender) nextChunkName() string {
	result := fmt.Sprintf("%d-%s", a.chunkId, a.appenderId)
	a.chunkId++
	return result
}
