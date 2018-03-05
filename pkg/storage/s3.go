package storage

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/net/context"
)

const ()

// newS3Backend returns a new backend implementation backed by the supplied
// S3 bucket name. Note that with the current interface, the AWS_REGION environment
// variable must be specified to use this backend.
func newS3Backend(bucketName string) (*s3Backend, error) {
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)

	return &s3Backend{
		bucketName: aws.String(bucketName),
		s3Client:   s3Client,
	}, nil
}

type s3Backend struct {
	bucketName *string
	s3Client   *s3.S3
}

func (b *s3Backend) read(ctx context.Context, id string) ([]byte, error) {
	getOutput, err := b.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: b.bucketName,
		Key:    aws.String(id),
	})
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(getOutput.Body)
}

func (b *s3Backend) write(ctx context.Context, id string, contents []byte) error {
	_, err := b.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: b.bucketName,
		Key:    aws.String(id),
		Body:   bytes.NewReader(contents),
	})

	return err
}

func (b *s3Backend) list(ctx context.Context, prefix string) ([]string, error) {
	listOutput, err := b.s3Client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: b.bucketName,
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return []string{}, err
	}

	var keys []string
	for _, obj := range listOutput.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

func (b *s3Backend) delete(ctx context.Context, id string) error {
	_, err := b.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: b.bucketName,
		Key:    aws.String(id),
	})

	return err
}
