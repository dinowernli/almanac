package appender

import (
	"sync"

	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type appender struct {
	entries            []*pb_almanac.LogEntry
	appendMutex        *sync.Mutex
	storage            storage.Storage
	maxEntriesPerChunk int64
}

// New returns a new appender backed by the supplied storage.
func New(storage storage.Storage, maxEntriesPerChunk int64) *appender {
	return &appender{
		storage:            storage,
		maxEntriesPerChunk: maxEntriesPerChunk,
		appendMutex:        &sync.Mutex{},
	}
}

func (a *appender) Append(ctx context.Context, request *pb_almanac.AppendRequest) (*pb_almanac.AppendResponse, error) {
	if request.Entry == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "no entry supplied")
	}

	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()
	a.entries = append(a.entries, request.Entry)

	// TODO(dino): If the chunk has grown enough, persist it to storage.

	return &pb_almanac.AppendResponse{}, nil
}
