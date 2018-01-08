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
	logger  *logrus.Logger
	storage *storage.Storage

	openChunks       []*openChunk
	openChunksMutex  *sync.Mutex
	closedChunksChan chan *openChunk

	maxChunkEntries    int
	maxChunkSpreadMs   int64
	maxChunkOpenTimeMs int64
}

// New returns a new appender backed by the supplied storage.
func New(logger *logrus.Logger, storage *storage.Storage, maxChunkEntries int, maxChunkSpreadMs int64, maxChunkOpenTimeMs int64) (*Appender, error) {
	if maxEntriesPerChunk < 1 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxEntriesPerChunk)
	}
	if maxChunkSpreadMs <= 0 {
		return nil, fmt.Errorf("must have positive chunk spread, but got: %d", maxChunkSpreadMs)
	}
	if maxChunkOpenTimeMs <= 0 {
		return nil, fmt.Errorf("must have positive max chunk open time, but got: %d", maxChunkOpenTimeMs)
	}

	result := &Appender{
		logger:  logger,
		storage: storage,

		openChunks:       []*openChunk{},
		openChunksMutex:  &sync.Mutex{},
		closedChunksChan: make(chan *openChunk),

		maxChunkEntries:    maxChunkEntries,
		maxChunkSpreadMs:   maxChunkSpreadMs,
		maxChunkOpenTimeMs: maxChunkOpenTimeMs,
	}

	// Kick of the background goroutine which sends closed chunks to storage.
	go result.storeClosedChunks()

	return result, nil
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
	entry := request.GetEntry()
	if entry == nil {
		err := grpc.Errorf(codes.InvalidArgument, "no entry supplied")
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	entryId := entry.GetId()
	if entryId == "" {
		err := grpc.Errorf(codes.InvalidArgument, "no id supplied")
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}
	logger = logger.WithFields(logrus.Fields{"entry": entryId})

	// Try to find an open chunk which can accept the entry.
	done := false
	for _, chunk := range a.openChunks {
		added, err := chunk.tryAdd(entry)
		if err != nil {
			err := grpc.Errorf(codes.Internal, "error while adding entry to chunk")
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}

		if added {
			done = true
			break
		}
	}

	// Open a new chunk if necessary.
	if !done {
		newChunk, err := newOpenChunk(entry, c.maxChunkEntries, c.maxChunkSpreadMs, c.maxChunkOpenTimeMs, c.closedChunksChan)
		if err != nil {
			err := grpc.Errorf(codes.Internal, "error while creating new chunk")
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
	}

	logger.Infof("Handled")
	return &pb_almanac.AppendResponse{}, nil
}

// storeClosedChunks takes all the chunk protos sent over the closed chunks
// channel and writes them to storage. This method blocks and is not expected
// to return for the lifetime of the appender, so it should be called in a
// dedicated goroutine.
func (a *Appender) storeClosedChunks() {
	for chunk := range a.closedChunksChan {
		chunkProto, err := chunk.toProto()
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to turn chunk into proto: %v", err)
			continue
		}

		err = a.storage.StoreChunk(chunkProto)
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to store chunk %v: %v", chunkProto.Id, err)
		}
	}
}
