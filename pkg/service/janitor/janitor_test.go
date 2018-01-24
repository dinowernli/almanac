package janitor

import (
	"testing"
	"time"

	st "dinowernli.me/almanac/pkg/storage"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	compactionInterval = 10 * time.Millisecond
	bigChunkMaxSpread  = 6 * time.Hour
)

func TestCompaction(t *testing.T) {
	storage, err := st.NewMemoryStorage()
	assert.NoError(t, err)

	_, err = New(context.Background(), logrus.New(), storage, compactionInterval, bigChunkMaxSpread)
	assert.NoError(t, err)

	// TODO(dino): Check that compaction has happened.
}
