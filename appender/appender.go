package appender

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type appender struct {
	entries     []*pb_almanac.LogEntry
	index       *index.Index
	appendMutex *sync.Mutex

	storage            storage.Storage
	maxEntriesPerChunk int
}

// New returns a new appender backed by the supplied storage.
func New(storage storage.Storage, maxEntriesPerChunk int) (*appender, error) {
	if maxEntriesPerChunk <= 0 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxEntriesPerChunk)
	}

	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}

	return &appender{
		index:              index,
		appendMutex:        &sync.Mutex{},
		storage:            storage,
		maxEntriesPerChunk: maxEntriesPerChunk,
	}, nil
}

func (a *appender) Append(ctx context.Context, request *pb_almanac.AppendRequest) (*pb_almanac.AppendResponse, error) {
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
		return nil, grpc.Errorf(codes.InvalidArgument, "unable to index raw jsen entry: %v", err)
	}

	a.entries = append(a.entries, request.Entry)

	// If the open chunk has grown enough, close it up and start a new one.
	if len(a.entries) >= a.maxEntriesPerChunk {
		err := a.storeChunk()
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "unable to store chunk: %v", err)
		}
	}

	return &pb_almanac.AppendResponse{}, nil
}

// storeChunk takes the currently open chunk, persists it to storage, and starts
// a new open chunk.
func (a *appender) storeChunk() error {
	indexProto, err := index.Serialize(a.index)
	if err != nil {
		return fmt.Errorf("unable to serialize index: %v", err)
	}

	chunkProto := &pb_almanac.Chunk{
		Entries: a.entries,
		Index:   indexProto,
	}

	// TODO(dino): Actually write this thing to storage.
	log.Printf("writing chunk to storage not implemented, chunkProto: %v\n", chunkProto)

	return nil
}
