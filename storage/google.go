package storage

import (
	"context"
	"encoding/pem"
	"io"
	"net/http"
	"time"

	"io/ioutil"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var (
	URLExpire = 10 * time.Minute
)

type GoogleCloudStorage struct {
	client     *storage.Client
	privateKey *pem.Block
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
	pemBlock, rest := pem.Decode(jwtConfig.PrivateKey)
	if len(rest) > 0 {
		return nil
	}

	return &GoogleCloudStorage{client: client, privateKey: pemBlock, accessID: accessID}
}

func (gcs *GoogleCloudStorage) Get(bucketName, repo, objectID string) (string, error) {
	return storage.SignedURL(bucketName, repo+"/"+objectID, &storage.SignedURLOptions{Method: http.MethodGet, PrivateKey: gcs.privateKey.Bytes, GoogleAccessID: gcs.AccessID, Expires: time.Now().Add(URLExpire)})
}

func (gcs *GoogleCloudStorage) GetObject(ctx context.Context, bucketName, repo, objectID string) (io.Reader, error) {
	return gcs.client.Bucket(bucketName).Object(repo + "/" + objectID).NewReader(ctx)
}

func (gcs *GoogleCloudStorage) Put(bucketName, repo, objectID string) (string, error) {
	return storage.SignedURL(bucketName, repo+"/"+objectID, &storage.SignedURLOptions{Method: http.MethodPut, PrivateKey: gcs.privateKey.Bytes, GoogleAccessID: gcs.AccessID, Expires: time.Now().Add(URLExpire)})
}

func (gcs *GoogleCloudStorage) PutObject(ctx context.Context, bucketName, repo, objectID string) (io.Writer, error) {
	return gcs.client.Bucket(bucketName).Object(repo + "/" + objectID).NewWriter(ctx), nil
}
