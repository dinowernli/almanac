package mixer

import (
	"container/heap"
	"fmt"
	"net/http"
	"time"

	almHttp "dinowernli.me/almanac/pkg/http"
	"dinowernli.me/almanac/pkg/service/discovery"
	"dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

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
	server.HandleFunc(httpUrl, m.handleHttp)
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
	for _, appender := range m.discovery.ListAppenders() {
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
		chunkIds, err := m.storage.ListChunks(request.StartMs, request.EndMs)
		if err != nil {
			return fmt.Errorf("unable to list chunks: %v", err)
		}
		for _, id := range chunkIds {
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

		// Incorporate the entry into our result set (including deduping).
		if _, ok := seen[entry.Id]; ok {
			continue
		}
		seen[entry.Id] = struct{}{}
		result = append(result, entry)
		if len(result) >= int(request.Num) {
			break
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

	ctx, _ := context.WithTimeout(context.Background(), httpSearchTimeoutMs*time.Millisecond)
	pageData.Response, pageData.Error = m.Search(ctx, pageData.Request)

	err := pageData.Render(writer)
	if err != nil {
		fmt.Fprintf(writer, "failed to render mixer page: %v", err)
	}
}
