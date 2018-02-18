package lfs

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/f110/git-lfs-cloud/config"
	"github.com/f110/git-lfs-cloud/storage"
	"golang.org/x/crypto/acme/autocert"
)

const (
	ContentType                           = "application/vnd.git-lfs+json"
	ErrorCodeNeedAuthenticationCredential = 401
	ErrorCodeForbidden                    = 403
	ErrorCodeNotExist                     = 404
	ErrorCodeNotAcceptable                = 406
	ErrorCodeRemoved                      = 410
	ErrorCodeValidation                   = 422
	ErrorCodeTooManyRequest               = 429
	ErrorCodeDiskFull                     = 507
	ErrorCodeBandwidthLimit               = 509
)

const (
	OperationWhoAmI   = "x-whoami"
	OperationDownload = "download"
	OperationUpload   = "upload"
)

type BatchRequest struct {
	Operation string            `json:"operation"`
	Transfers []string          `json:"transfers"`
	Refs      map[string]string `json:"refs"`
	Objects   []Object          `json:"objects"`
}

type BatchResponse struct {
	Transfer string   `json:"transfer"`
	Objects  []Object `json:"objects"`
}

type Object struct {
	Oid          string `json:"oid"`
	Size         int    `json:"size"`
	Autheticated bool   `json:"authenticated,omitempty"`
	Actions      Action `json:"actions,omitempty"`
	Erros        Error  `json:"errors,omitempty"`
}

type Action struct {
	Download Download `json:"download,omitempty"`
	Upload   Upload   `json:"upload,omitempty"`
	Verify   Verify   `json:"verify,omitempty"`
}

type Download struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt string            `json:"expires_at,omitempty"`
	ExpiresIn int64             `json:"expires_in,omitempty"`
}

type Upload Download
type Verify Download

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Server struct {
	Repositories map[string]repositoryConfig
}

type repositoryConfig struct {
	storageEngine storage.Storage
	bucketName    string
}

func NewServer(repositories map[string]*config.RepositoryConfig) *Server {
	reposConfig := make(map[string]repositoryConfig)
	for _, v := range repositories {
		var engine storage.Storage
		switch v.Storage {
		case "google":
			engine = storage.NewCloudStorage(v.AccessID, v.CredentialFile)
		case "nop":
			engine = &storage.Nop{}
		}
		reposConfig[v.Owner+"/"+v.Repo] = repositoryConfig{storageEngine: engine}
	}
	return &Server{Repositories: reposConfig}
}

func (server *Server) batchHandler(w http.ResponseWriter, req *http.Request) {
	splitedPath := strings.Split(req.URL.EscapedPath(), "/")[1:]
	repoName := ""
	for _, v := range splitedPath {
		if strings.Index(v, ".git") > 0 {
			repoName += v
			break
		}
		repoName += v + "/"
	}
	repoName = repoName[:strings.Index(repoName, ".git")]

	var batchReq BatchRequest
	var batchRes BatchResponse
	err := json.NewDecoder(req.Body).Decode(&batchReq)
	if err != nil {
		return
	}
	resObj := make([]Object, 0, len(batchReq.Objects))
	switch batchReq.Operation {
	case OperationDownload:
		for _, o := range batchReq.Objects {
			u := server.operationDownload(repoName, o.Oid)
			resObj = append(resObj, Object{
				Oid:          o.Oid,
				Size:         o.Size,
				Autheticated: true,
				Actions: Action{
					Download: Download{Href: u, ExpiresIn: time.Now().Add(5 * time.Minute).Unix()},
				},
			})
		}
	case OperationUpload:
		for _, o := range batchReq.Objects {
			u := server.operationUpload(repoName, o.Oid)
			resObj = append(resObj, Object{
				Oid:          o.Oid,
				Size:         o.Size,
				Autheticated: true,
				Actions: Action{
					Upload: Upload{Href: u, ExpiresIn: time.Now().Add(5 * time.Minute).Unix()},
				},
			})
		}
	}
	batchRes.Objects = resObj

	w.Header().Set("Content-Type", ContentType)
	buf, err := json.Marshal(batchRes)
	if err != nil {
		log.Print(err)
		return
	}
	w.Write(buf)
}

func (server *Server) operationDownload(repoName, objectID string) string {
	repoConf := server.Repositories[repoName]
	u, err := repoConf.storageEngine.Get(repoConf.bucketName, repoName, objectID)
	if err != nil {
		return ""
	}
	return u
}

func (server *Server) operationUpload(repoName, objectID string) string {
	repoConf := server.Repositories[repoName]
	u, err := repoConf.storageEngine.Put(repoConf.bucketName, repoName, objectID)
	if err != nil {
		return ""
	}
	return u
}

func (server *Server) ServeMux() http.Handler {
	m := &http.ServeMux{}
	m.HandleFunc("/", server.batchHandler)
	return m
}

func ObjectServer(disableHttps bool, cacheDir, host string, repos map[string]*config.RepositoryConfig) {
	serv := NewServer(repos)
	if disableHttps {
		s := &http.Server{
			Addr:    ":8080",
			Handler: serv.ServeMux(),
		}
		log.Println("starting lfs server on port 8080 (without TLS)...")
		log.Print(s.ListenAndServe())
	} else {
		m := &autocert.Manager{
			Cache:      autocert.DirCache(cacheDir),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(host),
		}
		go http.ListenAndServe(":http", m.HTTPHandler(nil))
		s := &http.Server{
			Addr:      ":https",
			Handler:   serv.ServeMux(),
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		log.Println("starting lfs server on port 443...")
		log.Print(s.ListenAndServeTLS("", ""))
	}
}
