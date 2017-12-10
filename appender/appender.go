package appender

import (
	"fmt"
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
	maxEntriesPerChunk int64
}

// New returns a new appender backed by the supplied storage.
func New(storage storage.Storage, maxEntriesPerChunk int64) (*appender, error) {
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
	logEntry := request.GetEntry()
	if logEntry == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "no entry supplied")
	}

	entryId := logEntry.GetId()
	if entryId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "no id supplied")
	}

	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()
	a.entries = append(a.entries, request.Entry)

	// TODO(dino): If the chunk has grown enough, persist it to storage.

	return &pb_almanac.AppendResponse{}, nil
}
