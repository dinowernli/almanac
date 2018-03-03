package http

import (
	"testing"

	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/stretchr/testify/assert"
)

func TestParseTimestamp(t *testing.T) {
	assert.Equal(t, int64(1234), ParseTimestamp("asdf", 1234))
	assert.Equal(t, int64(5678), ParseTimestamp("5678", 1234))
}

func TestRenderMixer(t *testing.T) {
	data := &MixerData{
		FormQuery:   "some query",
		FormStartMs: "1234",
		FormEndMs:   "5678",
		Error:       nil,
		Request:     &pb_almanac.SearchRequest{},
		Response:    &pb_almanac.SearchResponse{},
	}
	err := data.Render(&fakeWriter{})
	assert.NoError(t, err)
}

func TestRenderIngester(t *testing.T) {
	data := &IngesterData{
		FormContent: "some json blob",
		Error:       nil,
		Result:      "some result",
	}
	err := data.Render(&fakeWriter{})
	assert.NoError(t, err)
}

type fakeWriter struct {
}

func (w *fakeWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
