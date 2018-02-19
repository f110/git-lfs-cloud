package storage

import (
	"context"
	"io"

	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type AmazonS3 struct {
	client s3iface.S3API
}

func NewAmazonS3(region string) *AmazonS3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil
	}
	svs := s3.New(sess)

	return &AmazonS3{client: svs}
}

func (amazonS3 *AmazonS3) Get(bucketName string, repo string, objectID string) (string, error) {
	req, _ := amazonS3.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(repo + "/" + objectID),
	})
	u, err := req.Presign(URLExpire)
	if err != nil {
		return "", err
	}

	return u, nil
}

func (amazonS3 *AmazonS3) GetObject(ctx context.Context, bucketName string, repo string, objectID string) (io.ReadCloser, error) {
	res, err := amazonS3.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(repo + "/" + objectID),
	})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (amazonS3 *AmazonS3) Put(bucketName string, repo string, objectID string) (string, error) {
	req, _ := amazonS3.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(repo + "/" + objectID),
	})
	u, err := req.Presign(URLExpire)
	if err != nil {
		return "", err
	}

	return u, nil
}

func (amazonS3 *AmazonS3) PutObject(ctx context.Context, bucketName string, repo string, objectID string) (io.WriteCloser, error) {
	r, w := io.Pipe()
	buffer := bytes.NewBuffer([]byte{})
	go func() {
		io.Copy(buffer, r)

		buf := bytes.NewReader(buffer.Bytes())
		_, err := amazonS3.client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(repo + "/" + objectID),
			Body:   buf,
		})
		if err != nil {
			return
		}
	}()

	return w, nil
}
