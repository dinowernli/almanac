package janitor

import (
	"fmt"
	"time"

	st "dinowernli.me/almanac/pkg/storage"
	"dinowernli.me/almanac/pkg/util"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Janitor periodically takes a look at the contents in storage and may rewrite them to
// make queries cheaper and more efficient. This is intended to run as a singleton service.
type Janitor struct {
	ctx               context.Context
	logger            *logrus.Logger
	storage           *st.Storage
	cleanupInterval   time.Duration
	bigChunkMaxSpread time.Duration
}

// New creates a new Janitor instance which periodically compacts the supplied storage until
// the supplied context is done.
func New(ctx context.Context, logger *logrus.Logger, storage *st.Storage, cleanupInterval time.Duration, bigChunkMaxSpread time.Duration) (*Janitor, error) {
	if cleanupInterval <= 0 {
		return nil, fmt.Errorf("cleanup interval must be positive, but got %v", cleanupInterval)
	}
	if bigChunkMaxSpread <= 0 {
		return nil, fmt.Errorf("big chunk spread must be positive, but got: %v", bigChunkMaxSpread)
	}
	result := &Janitor{
		ctx:               ctx,
		logger:            logger,
		storage:           storage,
		cleanupInterval:   cleanupInterval,
		bigChunkMaxSpread: bigChunkMaxSpread,
	}
	result.start()
	return result, nil
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
	ctx, cancel := context.WithCancel(j.ctx)
	defer cancel()
	start := time.Now()

	chunks, err := j.storage.ListChunks(ctx, 0, 0, pb_almanac.ChunkId_SMALL)
	if err != nil {
		return fmt.Errorf("unable to list chunks during compaction: %v", err)
	}
	j.logger.Infof("Found %d small chunk(s) in storage", len(chunks))

	selectedChunkIds, err := j.selectSmallChunks(chunks)
	if err != nil {
		return fmt.Errorf("unable to select small chunks during compaction: %v", err)
	}
	if len(selectedChunkIds) == 0 {
		// Nothing to compact.
		return nil
	}
	j.logger.Infof("Selected %d small chunk(s) to compact", len(selectedChunkIds))

	bigChunk, err := j.constructBigChunk(j.ctx, selectedChunkIds)
	if err != nil {
		return fmt.Errorf("unable to construct big chunk during compaction: %v", err)
	}
	j.logger.Infof("Constructed big chunk with %d entries", len(bigChunk.Entries))

	_, err = j.storage.StoreChunk(j.ctx, bigChunk)
	if err != nil {
		return fmt.Errorf("unable to store big chunk during compaction: %v", err)
	}
	j.logger.Infof("Stored big chunk")

	err = j.deleteSmallChunks(j.ctx, selectedChunkIds)
	if err != nil {
		return fmt.Errorf("unable to delete small chunks during compaction: %v", err)
	}
	j.logger.Infof("Deleted %d small chunk(s) which have become redundant", len(selectedChunkIds))

	j.logger.Infof("Compaction successful, took %v", time.Since(start))
	return nil
}

// selectSmallChunks takes a list of id strings of small chunks and returns the ids which are to
// make up a new big chunk. All elements in the returned small chunks belong in the big chunk.
func (j *Janitor) selectSmallChunks(chunkIds []string) ([]*pb_almanac.ChunkId, error) {
	result := []*pb_almanac.ChunkId{}

	// Holds the start time of the big chunk we are writing.
	var startMs *int64

	// Holds the maximum end time of our new big chunk. This is used to determing when to stop
	// considering small chunks.
	var maxEndTime time.Time

	now := time.Now()
	for _, c := range chunkIds {
		idProto, err := st.ChunkIdProto(c)
		if err != nil {
			return nil, fmt.Errorf("unable to convert chunk %s id to proto: %v", c, err)
		}
		if idProto.Type != pb_almanac.ChunkId_SMALL {
			return nil, fmt.Errorf("expected small chunk but found type: %v", idProto.Type)
		}

		cStart := util.TimeMs(idProto.StartMs)
		cEnd := util.TimeMs(idProto.EndMs)

		// Only consider small chunks that have started sufficiently far in the past in
		// order to avoid creating big chunks for periods of time that are still actively
		// being written to.
		// TODO(dino): Introduce a separate duration rather than using bigChunkMaxSpread.
		if cStart.After(now.Add(-j.bigChunkMaxSpread)) {
			// Chunks are sorted by start, so no need to consider the next chunks.
			break
		}

		if startMs == nil {
			startMs = &idProto.StartMs
			maxEndTime = util.TimeMs(*startMs).Add(j.bigChunkMaxSpread)
		} else {
			// Stop once the current small chunk starts after the maximum end time of
			// the big chunk we are trying to build.
			if cStart.After(maxEndTime) {
				break
			}
		}

		// Even though the small chunk starts before our maximum end time, we don't want to
		// consider it unless it fits entirely inside the new big chunk.
		if cEnd.After(maxEndTime) {
			continue
		}

		result = append(result, idProto)
	}
	return result, nil
}

// constructBigChunk fetches all the data from the specified small chunks and returns a big chunk.
func (j *Janitor) constructBigChunk(ctx context.Context, smallChunkIds []*pb_almanac.ChunkId) (*pb_almanac.Chunk, error) {
	// TODO(dino): Parallelize this in a controlled way.
	allEntries := []*pb_almanac.LogEntry{}
	for _, c := range smallChunkIds {
		chunk, err := j.storage.LoadChunk(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("unable to load chunk %v: %v", c, err)
		}
		err = chunk.Close()
		if err != nil {
			return nil, fmt.Errorf("unable to close chunk: %v", err)
		}
		allEntries = append(allEntries, chunk.Entries()...)
	}

	chunk, err := st.ChunkProto(allEntries, pb_almanac.ChunkId_BIG)
	if err != nil {
		return nil, fmt.Errorf("unable to create large chunk: %v", err)
	}
	return chunk, nil
}

func (j *Janitor) deleteSmallChunks(ctx context.Context, smallChunkIds []*pb_almanac.ChunkId) error {
	// TODO(dino): Parallelize this in a controlled way.
	for _, c := range smallChunkIds {
		err := j.storage.DeleteChunk(ctx, c)
		if err != nil {
			return fmt.Errorf("failed to delete chunk: %v", err)
		}
	}
	return nil
}
