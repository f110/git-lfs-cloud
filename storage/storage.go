package storage

import (
	"bytes"
	"context"
	"io"
)

type Storage interface {
	Get(bucketName string, repo string, objectID string) (url string, err error)
	GetObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.Reader, err error)
	Put(bucketName string, repo string, objectID string) (url string, err error)
	PutObject(ctx context.Context, bucetName string, repo string, objectID string) (object io.Writer, err error)
}

type Nop struct{}

func (*Nop) Get(bucketName string, repo string, objectID string) (url string, err error) {
	return "http://example.com/lfs/objects/" + bucketName + "/" + repo + "/" + objectID, nil
}

func (*Nop) GetObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.Reader, err error) {
	return bytes.NewBuffer([]byte{}), nil
}

func (*Nop) Put(bucketName string, repo string, objectID string) (url string, err error) {
	return "https://example.com/lfs/objects/" + bucketName + "/" + repo + "/" + objectID, nil
}

func (*Nop) PutObject(ctx context.Context, bucetName string, repo string, objectID string) (object io.Writer, err error) {
	return bytes.NewBuffer([]byte{}), nil
}
