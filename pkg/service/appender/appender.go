package appender

import (
	"fmt"
	"sync"
	"time"

	"dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	// The number of random characters appended to chunk ids to make sure that
	// chunk ids are globally unique.
	chunkUidLength = 5

	// The time during which closed chunks are kept in memory even after writing
	// them out to storage. This should be longer than the typical time it takes
	// to serve a serach request on a mixer.
	closedChunkGracePeriodMs = 1000
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
	if maxChunkEntries < 1 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxChunkEntries)
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

	a.openChunksMutex.Lock()
	defer a.openChunksMutex.Unlock()

	results := []*pb_almanac.LogEntry{}
	for _, chunk := range a.openChunks {
		entries, err := chunk.search(ctx, request)
		if err != nil {
			err := fmt.Errorf("unable to search open chunk: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}

		for _, e := range entries {
			results = append(results, e)
		}
	}

	logger.Infof("Handled")
	return &pb_almanac.SearchResponse{Entries: results}, nil
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

	a.openChunksMutex.Lock()
	defer a.openChunksMutex.Unlock()

	// Try to find an open chunk which can accept the entry.
	done := false
	for _, chunk := range a.openChunks {
		added, err := chunk.tryAdd(entry)
		if err != nil {
			err := grpc.Errorf(codes.Internal, "error while adding entry to chunk: %v", err)
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
		newChunk, err := newOpenChunk(entry, a.maxChunkEntries, a.maxChunkSpreadMs, a.maxChunkOpenTimeMs, a.closedChunksChan)
		if err != nil {
			err := grpc.Errorf(codes.Internal, "error while creating new chunk: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
		a.openChunks = append(a.openChunks, newChunk)
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
		// Write the chunk out to storage.
		chunkProto, err := chunk.toProto()
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to turn chunk into proto: %v", err)
			continue
		}

		chunkId, err := a.storage.StoreChunk(context.TODO(), chunkProto)
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to store chunk %v: %v", chunkProto.Id, err)
		}

		// Now, remove it from the appender's list. We do this only after a grace period in order
		// to make sure that query mixers don't end up in the case where they hit storage *before*
		// the chunk is written, but hit this appender *after* the chunk is removed from memory.
		time.AfterFunc(time.Duration(closedChunkGracePeriodMs)*time.Millisecond, func() {
			a.removeOpenChunk(chunk)
		})

		a.logger.WithFields(logrus.Fields{"chunkId": chunkId}).Infof("Stored chunk with %d entries", len(chunkProto.Entries))
	}
}

// removeOpenChunk removes the supplied chunk from the list of open chunks.
func (a *Appender) removeOpenChunk(chunk *openChunk) {
	a.openChunksMutex.Lock()
	defer a.openChunksMutex.Unlock()

	newOpenChunks := []*openChunk{}
	for _, c := range a.openChunks {
		if c != chunk {
			newOpenChunks = append(newOpenChunks, c)
		}
	}
	a.openChunks = newOpenChunks
}
