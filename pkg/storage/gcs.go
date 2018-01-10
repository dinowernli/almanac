package storage

import (
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// NewGcsBackend returns a new backend implementation backed by the supplied
// gcs bucket name.
func NewGcsBackend(bucketName string) (*gcsBackend, error) {
	gcsClient, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to create gcs client: %v", err)
	}
	return &gcsBackend{bucket: gcsClient.Bucket(bucketName)}, nil
}

type gcsBackend struct {
	bucket *storage.BucketHandle
}

func (b *gcsBackend) read(id string) ([]byte, error) {
	r, err := b.bucket.Object(id).NewReader(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to open reader for object %s: %v", id, err)
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func (b *gcsBackend) write(id string, contents []byte) error {
	w := b.bucket.Object(id).NewWriter(context.TODO())
	defer w.Close()

	_, err := w.Write(contents)
	if err != nil {
		return fmt.Errorf("unable to write to object %s: %v", id, err)
	}
	return nil
}

func (b *gcsBackend) list(prefix string) ([]string, error) {
	it := b.bucket.Objects(context.TODO(), &storage.Query{Prefix: prefix})

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
