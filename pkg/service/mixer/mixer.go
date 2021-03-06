package mixer

import (
	"container/heap"
	"fmt"
	"net/http"
	"time"

	almHttp "github.com/dinowernli/almanac/pkg/http"
	"github.com/dinowernli/almanac/pkg/service/discovery"
	"github.com/dinowernli/almanac/pkg/storage"
	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	httpUrl             = "/mixer"
	urlParamQuery       = "q"
	urlParamStartMs     = "s"
	urlParamEndMs       = "e"
	httpSearchTimeoutMs = 3000
)

var (
	searchField = logrus.Fields{"method": "mixer.Search"}
)

// Mixer is an implementation of the mixer rpc service. It provides global
// search functionality across the entire system.
type Mixer struct {
	logger    *logrus.Logger
	storage   *storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(logger *logrus.Logger, storage *storage.Storage, discovery *discovery.Discovery) *Mixer {
	return &Mixer{logger: logger, storage: storage, discovery: discovery}
}

// RegisterHttp registers a page on the supplied server, used for executing searches.
func (m *Mixer) RegisterHttp(server *http.ServeMux) {
	server.HandleFunc(httpUrl, prometheus.InstrumentHandlerFunc(httpUrl, m.handleHttp))
}

func (m *Mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	logger := m.logger.WithFields(searchField)
	if request.StartMs != 0 && request.EndMs != 0 && request.StartMs > request.EndMs {
		err := grpc.Errorf(codes.InvalidArgument, "cannot query, start(%d) is greater than end (%d)", request.StartMs, request.EndMs)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	searchHeap := &searchHeap{}
	heap.Init(searchHeap)
	g, _ := errgroup.WithContext(ctx)

	// Compute one heap item for every appender.
	for _, a := range m.discovery.ListAppenders() {
		// Copy the iteration variable here because otherwise, all instances of the func below end up
		// using the same appender because golang loop variables are by reference.
		appender := a
		g.Go(func() error {
			response, err := appender.Search(ctx, request)
			if err != nil {
				return fmt.Errorf("unable to search appender: %v", err)
			}
			if len(response.Entries) > 0 {
				heap.Push(searchHeap, &appenderHeapItem{entries: response.Entries, idx: 0})
			}
			return nil
		})
	}

	// Compute a heap item for every chunk whose time span overlaps with our query.
	g.Go(func() error {
		smallChunkIds, err := m.storage.ListChunks(ctx, request.StartMs, request.EndMs, pb_almanac.ChunkId_SMALL)
		if err != nil {
			return fmt.Errorf("unable to list small chunks: %v", err)
		}
		bigChunkIds, err := m.storage.ListChunks(ctx, request.StartMs, request.EndMs, pb_almanac.ChunkId_BIG)
		if err != nil {
			return fmt.Errorf("unable to list big chunks: %v", err)
		}
		allChunkIds := append(smallChunkIds, bigChunkIds...)

		for _, id := range allChunkIds {
			idProto, err := storage.ChunkIdProto(id)
			if err != nil {
				return fmt.Errorf("unable to compute chunk id proto: %v", err)
			}
			heap.Push(searchHeap, &chunkHeapItem{chunkIdProto: idProto, searchRequest: request, ctx: ctx, storage: m.storage})
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		err := grpc.Errorf(codes.Internal, "search failed: %v", err)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	// Start assembling results by repeatedly grabbing the next one from the heap.
	result := []*pb_almanac.LogEntry{}
	seen := map[string]struct{}{}
	for searchHeap.Len() > 0 {
		item := heap.Pop(searchHeap).(heapItem)
		entry, err := item.entry()
		if err != nil {
			err := grpc.Errorf(codes.Internal, "unable to extract entry from heap item: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
		if entry == nil {
			// This can happen if the entire chunk has no entries to offer at all. Abandon this chunk.
			continue
		}

		// Incorporate the entry into our result set (including deduping).
		if _, ok := seen[entry.Id]; !ok {
			seen[entry.Id] = struct{}{}
			result = append(result, entry)
			if len(result) >= int(request.Num) {
				break
			}
		}

		// Re-populate the heap if necessary.
		next, err := item.next()
		if err != nil {
			err := grpc.Errorf(codes.Internal, "unable to extract next item from heap item: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
		if next != nil {
			heap.Push(searchHeap, next)
		}
	}

	logger = logger.WithFields(logrus.Fields{"hits": len(result)})
	logger.Infof("Handled")
	return &pb_almanac.SearchResponse{Entries: result}, nil
}

// handleHttp serves a web page which can be used to execute queries on this mixer.
func (m *Mixer) handleHttp(writer http.ResponseWriter, request *http.Request) {
	pageData := &almHttp.MixerData{
		FormQuery:   request.FormValue(urlParamQuery),
		FormStartMs: request.FormValue(urlParamStartMs),
		FormEndMs:   request.FormValue(urlParamEndMs),
	}

	if pageData.FormQuery == "" && pageData.FormStartMs == "" && pageData.FormEndMs == "" {
		err := pageData.Render(writer)
		if err != nil {
			fmt.Fprintf(writer, "failed to render empty mixer page: %v", err)
		}
		return
	}

	pageData.Request = &pb_almanac.SearchRequest{
		Query:   pageData.FormQuery,
		Num:     100,
		StartMs: almHttp.ParseTimestamp(pageData.FormStartMs, 0),
		EndMs:   almHttp.ParseTimestamp(pageData.FormEndMs, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpSearchTimeoutMs*time.Millisecond)
	defer cancel()
	pageData.Response, pageData.Error = m.Search(ctx, pageData.Request)

	err := pageData.Render(writer)
	if err != nil {
		fmt.Fprintf(writer, "failed to render mixer page: %v", err)
	}
}
