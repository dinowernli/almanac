package storage

import (
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const (
	gcsClientTimeoutMs = 3000
)

// NewGcsBackend returns a new backend implementation backed by the supplied
// gcs bucket name.
func NewGcsBackend(bucketName string) (*gcsBackend, error) {
	ctx, _ := context.WithTimeout(context.Background(), gcsClientTimeoutMs*time.Millisecond)
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create gcs client: %v", err)
	}
	return &gcsBackend{bucket: gcsClient.Bucket(bucketName)}, nil
}

type gcsBackend struct {
	bucket *storage.BucketHandle
}

func (b *gcsBackend) read(id string) ([]byte, error) {
	return nil, fmt.Errorf("read not implemented")
}

func (b *gcsBackend) write(id string, contents []byte) error {
	return fmt.Errorf("write not implemented")
}

func (b *gcsBackend) list(prefix string) ([]string, error) {
	return nil, fmt.Errorf("list not implemented")
}
