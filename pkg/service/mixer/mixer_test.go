package mixer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dinowernli/almanac/pkg/service/discovery"
	st "github.com/dinowernli/almanac/pkg/storage"
	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	entry1 = &pb_almanac.LogEntry{
		Id:          "id1",
		EntryJson:   `{ "message": "foo" }`,
		TimestampMs: int64(100),
	}
)

func TestMixerCallsAppenders(t *testing.T) {
	storage, err := st.NewMemoryStorage()
	assert.NoError(t, err)

	appenders := []pb_almanac.AppenderClient{&fakeAppender{}, &fakeAppender{}, &fakeAppender{}, &fakeAppender{}}
	mixer := New(logrus.New(), storage, discovery.NewForTesting(appenders))

	_, err = mixer.Search(context.Background(), &pb_almanac.SearchRequest{})
	assert.NoError(t, err)

	for _, appender := range appenders {
		assert.Equal(t, 1, appender.(*fakeAppender).searchCalls)
	}
}

func TestHttp(t *testing.T) {
	storage, err := st.NewMemoryStorage()
	assert.NoError(t, err)

	appenders := []pb_almanac.AppenderClient{&fakeAppender{}}
	mixer := New(logrus.New(), storage, discovery.NewForTesting(appenders))

	request, err := http.NewRequest("GET", "/mixer", nil)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	mixer.handleHttp(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Cheap way to check that the template rendered correctly.
	assert.True(t, strings.Contains(recorder.Body.String(), "Mixer"))
}

func TestSearchNoResults(t *testing.T) {
	chunk, err := st.ChunkProto([]*pb_almanac.LogEntry{entry1}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)

	storage, err := st.NewMemoryStorage()
	assert.NoError(t, err)

	_, err = storage.StoreChunk(context.Background(), chunk)
	assert.NoError(t, err)

	appenders := []pb_almanac.AppenderClient{&fakeAppender{}}
	mixer := New(logrus.New(), storage, discovery.NewForTesting(appenders))

	// Search for a different term than what's in the chunk, such that the chunk gets searched
	// but returns no results.
	response, err := mixer.Search(context.Background(), &pb_almanac.SearchRequest{
		StartMs: 1,
		EndMs:   1000,
		Query:   "baz",
		Num:     100,
	})
	assert.NoError(t, err)
	assert.Empty(t, response.Entries)
}

type fakeAppender struct {
	searchCalls int
}

func (a *fakeAppender) Search(ctx context.Context, request *pb_almanac.SearchRequest, options ...grpc.CallOption) (*pb_almanac.SearchResponse, error) {
	a.searchCalls++
	return &pb_almanac.SearchResponse{}, nil
}

func (a *fakeAppender) Append(ctx context.Context, request *pb_almanac.AppendRequest, options ...grpc.CallOption) (*pb_almanac.AppendResponse, error) {
	return &pb_almanac.AppendResponse{}, nil
}
