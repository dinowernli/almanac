package appender

import (
	"encoding/json"
	"fmt"
	"sync"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	chunkUidLength = 5
)

var (
	appendField = logrus.Fields{"method": "appender.Append"}
	searchField = logrus.Fields{"method": "appender.Search"}
)

// Appender keeps track of a single open chunk under construction at a time and
// periodically decides that the open chunk is complete, at which point the
// appender writes the chunk to storage. The appender also knows how to answer
// search requests for the currently open chunk.
type Appender struct {
	logger *logrus.Logger

	entries map[string]*pb_almanac.LogEntry
	index   *index.Index
	chunkId *pb_almanac.ChunkId

	openChunks []*openChunk

	appendMutex *sync.Mutex

	storage            *storage.Storage
	maxEntriesPerChunk int
}

// New returns a new appender backed by the supplied storage.
func New(logger *logrus.Logger, storage *storage.Storage, maxEntriesPerChunk int) (*Appender, error) {
	if maxEntriesPerChunk < 1 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxEntriesPerChunk)
	}

	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}

	return &Appender{
		logger: logger,

		// TODO(dino): Remove these.
		entries: map[string]*pb_almanac.LogEntry{},
		index:   index,
		chunkId: newChunkId(int64(3434)),

		appendMutex:        &sync.Mutex{},
		storage:            storage,
		maxEntriesPerChunk: maxEntriesPerChunk,
	}, nil
}

func (a *Appender) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	logger := a.logger.WithFields(searchField)

	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	ids, err := a.index.Search(ctx, request.Query, request.Num)
	if err != nil {
		err := fmt.Errorf("unable to search index: %v", err)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	entries := []*pb_almanac.LogEntry{}
	for _, id := range ids {
		entry, ok := a.entries[id]
		if !ok {
			err := fmt.Errorf("could not locate hit %s", id)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}

		if request.StartMs != 0 && entry.TimestampMs < request.StartMs {
			continue
		}
		if request.EndMs != 0 && entry.TimestampMs > request.EndMs {
			continue
		}
		entries = append(entries, entry)
	}

	logger.Infof("Handled")
	return &pb_almanac.SearchResponse{Entries: entries}, nil
}

func (a *Appender) Append(ctx context.Context, request *pb_almanac.AppendRequest) (*pb_almanac.AppendResponse, error) {
	logger := a.logger.WithFields(appendField)

	// Perform some validation of the request.
	logEntry := request.GetEntry()
	if logEntry == nil {
		err := grpc.Errorf(codes.InvalidArgument, "no entry supplied")
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	entryId := logEntry.GetId()
	if entryId == "" {
		err := grpc.Errorf(codes.InvalidArgument, "no id supplied")
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}
	logger = logger.WithFields(logrus.Fields{"entry": entryId})

	// Update the data structures corresponding to the open chunk.
	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	var rawEntry interface{}
	err := json.Unmarshal([]byte(logEntry.EntryJson), &rawEntry)
	if err != nil {
		err := grpc.Errorf(codes.InvalidArgument, "unable to parse raw json: %v", err)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	err = a.index.Index(entryId, rawEntry)
	if err != nil {
		err := grpc.Errorf(codes.Internal, "unable to index raw json entry: %v", err)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	a.entries[entryId] = request.Entry
	ts := request.Entry.TimestampMs
	if ts > a.chunkId.EndMs {
		a.chunkId.EndMs = ts
	}
	if ts < a.chunkId.StartMs {
		a.chunkId.StartMs = ts
	}

	// If the open chunk has grown enough, close it up and start a new one.
	if len(a.entries) >= a.maxEntriesPerChunk {
		err := a.storeChunk()
		if err != nil {
			err := grpc.Errorf(codes.Internal, "unable to store chunk: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}

		index, err := index.NewIndex()
		if err != nil {
			// TODO(dino): If this happens, stop responding to requests.
			err := grpc.Errorf(codes.Internal, "unable to create new index: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
		a.index = index
		a.entries = map[string]*pb_almanac.LogEntry{}
		a.chunkId = newChunkId(int64(3434))
	}

	logger.Infof("Handled")
	return &pb_almanac.AppendResponse{}, nil
}

// storeChunk takes the currently open chunk, persists it to storage, and starts
// a new open chunk.
func (a *Appender) storeChunk() error {
	indexProto, err := index.Serialize(a.index)
	if err != nil {
		return fmt.Errorf("unable to serialize index: %v", err)
	}
	a.index.Close()

	entries := []*pb_almanac.LogEntry{}
	for _, e := range a.entries {
		entries = append(entries, e)
	}

	chunkProto := &pb_almanac.Chunk{
		Id:      a.chunkId,
		Entries: entries,
		Index:   indexProto,
	}

	err = a.storage.StoreChunk(chunkProto)
	if err != nil {
		return fmt.Errorf("unable to store chunk %v: %v", chunkProto.Id, err)
	}
	return nil
}
