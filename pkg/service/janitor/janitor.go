package janitor

import (
	"fmt"
	"time"

	st "dinowernli.me/almanac/pkg/storage"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Janitor periodically takes a look at the contents in storage and may rewrite them to
// make queries cheaper and more efficient. This is intended to run as a singleton service.
type Janitor struct {
	ctx             context.Context
	logger          *logrus.Logger
	storage         *st.Storage
	cleanupInterval time.Duration
}

// New creates a new Janitor instance which periodically compacts the supplied storage until
// the supplied context is done.
func New(ctx context.Context, logger *logrus.Logger, storage *st.Storage, cleanupInterval time.Duration) *Janitor {
	result := &Janitor{
		ctx:             ctx,
		logger:          logger,
		storage:         storage,
		cleanupInterval: cleanupInterval,
	}
	result.start()
	return result
}

func (j *Janitor) start() {
	ticker := time.NewTicker(j.cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := j.executeCompaction()
				if err != nil {
					j.logger.WithError(err).Warn("Compaction failed")
				}
			case <-j.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (j *Janitor) executeCompaction() error {
	return fmt.Errorf("compaction not implemented")
}
