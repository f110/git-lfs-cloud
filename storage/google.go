package storage

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type GoogleCloudStorage struct {
	client     *storage.Client
	privateKey []byte
	accessID   string
}

func NewCloudStorage(accessID, credentialFile string) *GoogleCloudStorage {
	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(credentialFile))
	if err != nil {
		return nil
	}
	creds, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		return nil
	}
	jwtConfig, err := google.JWTConfigFromJSON(creds, "")
	if err != nil {
		return nil
	}

	return &GoogleCloudStorage{client: client, privateKey: jwtConfig.PrivateKey, accessID: jwtConfig.Email}
}

func (gcs *GoogleCloudStorage) Get(bucketName, repo, objectID string) (string, error) {
	return storage.SignedURL(bucketName, repo+"/"+objectID, &storage.SignedURLOptions{
		Method:         http.MethodGet,
		PrivateKey:     gcs.privateKey,
		GoogleAccessID: gcs.accessID,
		Expires:        time.Now().Add(URLExpire),
		ContentType:    "application/octet-stream",
	})
}

func (gcs *GoogleCloudStorage) GetObject(ctx context.Context, bucketName, repo, objectID string) (io.ReadCloser, error) {
	return gcs.client.Bucket(bucketName).Object(repo + "/" + objectID).NewReader(ctx)
}

func (gcs *GoogleCloudStorage) Put(bucketName, repo, objectID string) (string, error) {
	return storage.SignedURL(bucketName, repo+"/"+objectID, &storage.SignedURLOptions{
		Method:         http.MethodPut,
		PrivateKey:     gcs.privateKey,
		GoogleAccessID: gcs.accessID,
		Expires:        time.Now().Add(URLExpire),
		ContentType:    "application/octet-stream",
	},
	)
}

func (gcs *GoogleCloudStorage) PutObject(ctx context.Context, bucketName, repo, objectID string) (io.WriteCloser, error) {
	return gcs.client.Bucket(bucketName).Object(repo + "/" + objectID).NewWriter(ctx), nil
}
