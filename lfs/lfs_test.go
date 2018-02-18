package lfs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/f110/git-lfs-cloud/config"
)

func TestServer(t *testing.T) {
	serv := NewServer(map[string]*config.RepositoryConfig{"f110/test1": {Owner: "f110", Repo: "test1", Storage: "nop"}})
	s := httptest.NewServer(serv.ServeMux())

	t.Run("batchHandler_download", func(t *testing.T) {
		batchReq := &BatchRequest{
			Operation: OperationDownload,
			Transfers: []string{"basic"},
			Refs:      map[string]string{"name": "refs/head/master"},
			Objects:   []Object{{Oid: "12345678", Size: 123}},
		}
		reqBody, err := json.Marshal(batchReq)
		if err != nil {
			t.Error(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL+"/f110/test1.git/info/lfs/objects/batch", bytes.NewReader(reqBody))
		if err != nil {
			t.Error(err)
		}
		req.Header.Add("Content-Type", ContentType)
		req.Header.Add("Accept", ContentType)
		req.Header.Add("Authorization", "Bearer for-test")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.Header.Get("Content-Type") != ContentType {
			t.Errorf("content-type is not %s. %s", ContentType, res.Header.Get("Content-Type"))
		}

		var batchRes BatchResponse
		err = json.NewDecoder(res.Body).Decode(&batchRes)
		if err != nil {
			t.Fatal(err)
		}

		if len(batchRes.Objects) != 1 {
			t.Fatalf("Response: objects length is mismatch: %d", len(batchRes.Objects))
		}
		if batchRes.Objects[0].Oid != "12345678" {
			t.Error("Oid is mismatch")
		}
		if batchRes.Objects[0].Size != 123 {
			t.Error("Size is mismatch")
		}
		if batchRes.Objects[0].Actions.Download.Href == "" {
			t.Error("Response: download url is not found")
		}
		if batchRes.Objects[0].Actions.Download.ExpiresIn < time.Now().Add(1*time.Minute).Unix() {
			t.Error("Response: expires is too short or not present")
		}
	})

	t.Run("batchHandler_upload", func(t *testing.T) {
		batchReq := &BatchRequest{
			Operation: OperationUpload,
			Transfers: []string{"basic"},
			Refs:      map[string]string{"name": "refs/head/master"},
			Objects:   []Object{{Oid: "12345678", Size: 123}},
		}
		reqBody, err := json.Marshal(batchReq)
		if err != nil {
			t.Error(err)
		}

		req, err := http.NewRequest(http.MethodPost, s.URL+"/f110/test1.git/info/lfs/objects/batch", bytes.NewReader(reqBody))
		if err != nil {
			t.Error(err)
		}
		req.Header.Add("Content-Type", ContentType)
		req.Header.Add("Accept", ContentType)
		req.Header.Add("Authorization", "Bearer for-test")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if res.Header.Get("Content-Type") != ContentType {
			t.Errorf("content-type is not %s. %s", ContentType, res.Header.Get("Content-Type"))
		}

		var batchRes BatchResponse
		err = json.NewDecoder(res.Body).Decode(&batchRes)
		if err != nil {
			t.Fatal(err)
		}

		if len(batchRes.Objects) != 1 {
			t.Fatalf("Response: objects length is mismatch: %d", len(batchRes.Objects))
		}
		if batchRes.Objects[0].Oid != "12345678" {
			t.Error("Oid is mismatch")
		}
		if batchRes.Objects[0].Size != 123 {
			t.Error("Size is mismatch")
		}
		if batchRes.Objects[0].Actions.Upload.Href == "" {
			t.Error("Response: download url is not found")
		}
		if batchRes.Objects[0].Actions.Upload.ExpiresIn < time.Now().Add(1*time.Minute).Unix() {
			t.Error("Response: expires is too short or not present")
		}
	})

	s.Close()
}
