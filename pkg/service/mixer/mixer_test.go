package mixer

import (
	"testing"

	"dinowernli.me/almanac/pkg/service/discovery"
	st "dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestMixerCallsAppenders(t *testing.T) {
	appenders := []pb_almanac.AppenderClient{&fakeAppender{}, &fakeAppender{}, &fakeAppender{}, &fakeAppender{}}
	mixer := New(logrus.New(), st.NewInMemoryStorage(), discovery.NewForTesting(appenders))

	_, err := mixer.Search(context.Background(), &pb_almanac.SearchRequest{})
	assert.NoError(t, err)

	for _, appender := range appenders {
		assert.Equal(t, 1, appender.(*fakeAppender).searchCalls)
	}
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
