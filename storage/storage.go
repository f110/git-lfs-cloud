package storage

import (
	"bytes"
	"context"
	"io"
	"time"
)

var (
	URLExpire = 10 * time.Minute
)

type Storage interface {
	Get(bucketName string, repo string, objectID string) (url string, err error)
	GetObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.ReadCloser, err error)
	Put(bucketName string, repo string, objectID string) (url string, err error)
	PutObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.WriteCloser, err error)
}

type Nop struct{}

func (*Nop) Get(bucketName string, repo string, objectID string) (url string, err error) {
	return "http://example.com/lfs/objects/" + bucketName + "/" + repo + "/" + objectID, nil
}

func (*Nop) GetObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.ReadCloser, err error) {
	return bytes.NewBuffer([]byte{}), nil
}

func (*Nop) Put(bucketName string, repo string, objectID string) (url string, err error) {
	return "https://example.com/lfs/objects/" + bucketName + "/" + repo + "/" + objectID, nil
}

func (*Nop) PutObject(ctx context.Context, bucketName string, repo string, objectID string) (object io.WriteCloser, err error) {
	return bytes.NewBuffer([]byte{}), nil
}
