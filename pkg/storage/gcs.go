package storage

import (
	"fmt"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const (
	gcsReadTimeout   = 1 * time.Second
	gcsWriteTimeout  = 1 * time.Second
	gcsListTimeout   = 1 * time.Second
	gcsDeleteTimeout = 1 * time.Second
)

// newGcsBackend returns a new backend implementation backed by the supplied
// gcs bucket name.
func newGcsBackend(bucketName string) (*gcsBackend, error) {
	gcsClient, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to create gcs client: %v", err)
	}
	return &gcsBackend{bucket: gcsClient.Bucket(bucketName)}, nil
}

type gcsBackend struct {
	bucket *storage.BucketHandle
}

func (b *gcsBackend) read(ctx context.Context, id string) ([]byte, error) {
	c, f := context.WithTimeout(ctx, gcsReadTimeout)
	defer f()

	r, err := b.bucket.Object(id).NewReader(c)
	if err != nil {
		return nil, fmt.Errorf("unable to open reader for object %s: %v", id, err)
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func (b *gcsBackend) write(ctx context.Context, id string, contents []byte) error {
	c, f := context.WithTimeout(ctx, gcsWriteTimeout)
	defer f()

	w := b.bucket.Object(id).NewWriter(c)
	defer w.Close()

	_, err := w.Write(contents)
	if err != nil {
		return fmt.Errorf("unable to write to object %s: %v", id, err)
	}
	return nil
}

func (b *gcsBackend) list(ctx context.Context, prefix string) ([]string, error) {
	c, f := context.WithTimeout(ctx, gcsListTimeout)
	defer f()

	it := b.bucket.Objects(c, &storage.Query{Prefix: prefix})
	result := []string{}
	for {
		attributes, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to list objects with prefix %s: %v", prefix, err)
		}
		result = append(result, attributes.Name)
	}
	return result, nil
}

func (b *gcsBackend) delete(ctx context.Context, id string) error {
	c, f := context.WithTimeout(ctx, gcsDeleteTimeout)
	defer f()

	err := b.bucket.Object(id).Delete(c)
	if err != nil {
		return fmt.Errorf("gcs request to delete %s failed: %v", id, err)
	}
	return nil
}
