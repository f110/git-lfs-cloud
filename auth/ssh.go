package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/f110/git-lfs-cloud/database"
	"github.com/f110/git-lfs-cloud/lfs"
	"github.com/gliderlabs/ssh"
)

const (
	HostKeyBits         = 4096
	TokenExpire         = 3600 // 1 hour
	AuthenticateCommand = "git-lfs-authenticate"
)

type Authenticate struct {
	Header    map[string]string `json:"header"`
	ExpiresIn int               `json:"expires_in,omitempty"`
	ExpiresAt int               `json:"expires_at,omitempty"`
}

func generateHostKey() ([]byte, error) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, HostKeyBits)
	pb := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	b := pem.EncodeToMemory(pb)
	return b, nil
}

func readOrGenerateHostKey() ([]byte, error) {
	hostKey, err := database.Credential{}.ReadHostKey()
	if err == database.ErrBucketNotFound {
		log.Println("Host key not found. generate new host key.")
		newHostKey, err := generateHostKey()
		if err != nil {
			return nil, err
		}
		err = database.Credential{}.SaveHostKey(newHostKey)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		hostKey = newHostKey
	}
	return hostKey, nil
}

func handleWhoAmI(session ssh.Session, user string, _repo string) {
	io.WriteString(session, fmt.Sprintf("Hi %s!\n", user))
}

func handleDownload(session ssh.Session, user, repo string) {
	sess, err := lfs.FindSession(user)
	if err != nil {
		newSession, err := lfs.NewSession(user)
		if err != nil {
			io.WriteString(session, "failed authorization\n")
			return
		}
		sess = newSession
	}
	auth := &Authenticate{
		Header:    map[string]string{"Authorization": "Bearer " + sess.ID},
		ExpiresIn: int(time.Now().Unix() + TokenExpire),
	}
	buf, err := json.Marshal(auth)
	if err != nil {
		io.WriteString(session, "failed authorization\n")
		return
	}
	io.WriteString(session, string(buf))
}

func handleUpload(session ssh.Session, user, repo string) {
	handleDownload(session, user, repo)
}

func SSHServer() {
	hostKey, err := readOrGenerateHostKey()
	if err != nil {
		log.Print(err)
		return
	}
	hostKeyOption := ssh.HostKeyPEM(hostKey)

	ssh.Handle(func(s ssh.Session) {
		var repo string
		var operation string
		var username string

		if s.Command()[0] == AuthenticateCommand {
			repo = s.Command()[1]
			operation = s.Command()[2]
		} else {
			io.WriteString(s, "not supported\n")
			return
		}

		pubKey := s.PublicKey()
	FindUserLoop:
		for user, pubKeys := range PermitPublicKeys {
			for _, pub := range pubKeys {
				if ssh.KeysEqual(pubKey, pub) {
					username = user
					break FindUserLoop
				}
			}
		}

		switch operation {
		case lfs.OperationWhoAmI:
			handleWhoAmI(s, username, repo)
		case lfs.OperationDownload:
			handleDownload(s, username, repo)
		case lfs.OperationUpload:
			handleUpload(s, username, repo)
		default:
			io.WriteString(s, "not supported operation")
		}
	})

	publicKeyOption := ssh.PublicKeyAuth(func(user string, key ssh.PublicKey) bool {
		for _, pubKeys := range PermitPublicKeys {
			for _, pubKey := range pubKeys {
				if ssh.KeysEqual(key, pubKey) {
					return true
				}
			}
		}
		return false
	})

	log.Print("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil, hostKeyOption, publicKeyOption))
}
