package database

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gliderlabs/ssh"
)

var (
	Conn *bolt.DB
)

var (
	BucketCredential      = []byte("Credential")
	BucketPublicKeys      = []byte("PublicKeys")
	BucketRepositoryUsers = []byte("RepoUsers")
	KeyHostKey            = []byte("HostKey")
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrNotFound       = errors.New("not found")
)

type UserPublicKeys struct {
	Username   string
	PublicKeys [][]byte
	UpdatedAt  time.Time
}

type Credential struct{}

func SaveRepositoryUsers(repo string, users []string) error {
	return Conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketRepositoryUsers)
		if b == nil {
			newBucket, err := tx.CreateBucketIfNotExists(BucketRepositoryUsers)
			if err != nil {
				return err
			}
			b = newBucket
		}
		return b.Put([]byte(repo), []byte(strings.Join(users, ",")))
	})
}

func ReadRepositoryUsers(repo string) ([]string, error) {
	tx, err := Conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(BucketRepositoryUsers)
	if b == nil {
		return nil, ErrBucketNotFound
	}

	buf := b.Get([]byte(repo))
	if len(buf) == 0 {
		return nil, ErrNotFound
	}
	users := strings.Split(string(buf), ",")
	return users, nil
}

func (Credential) SaveHostKey(hostKey []byte) error {
	return Conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCredential)
		if b == nil {
			newBucket, err := tx.CreateBucket(BucketCredential)
			if err != nil {
				return err
			}
			b = newBucket
		}

		return b.Put(KeyHostKey, hostKey)
	})
}

func (Credential) ReadHostKey() ([]byte, error) {
	tx, err := Conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(BucketCredential)
	if b == nil {
		return nil, ErrBucketNotFound
	}
	return b.Get(KeyHostKey), nil
}

func SavePubKey(name string, pubKeys []ssh.PublicKey) error {
	keys := &UserPublicKeys{Username: name}
	for _, v := range pubKeys {
		keys.PublicKeys = append(keys.PublicKeys, v.Marshal())
	}
	value, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	return Conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketPublicKeys)
		if b == nil {
			newBucket, err := tx.CreateBucketIfNotExists(BucketPublicKeys)
			if err != nil {
				return err
			}
			b = newBucket
		}
		return b.Put([]byte(name), value)
	})
}

func ReadPublicKeys(name string) ([]ssh.PublicKey, error) {
	tx, err := Conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(BucketPublicKeys)
	if b == nil {
		return nil, ErrBucketNotFound
	}

	buf := b.Get([]byte(name))
	if buf == nil {
		return nil, ErrNotFound
	}
	keys := &UserPublicKeys{UpdatedAt: time.Now()}
	err = json.Unmarshal(buf, keys)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]ssh.PublicKey, 0)
	for _, key := range keys.PublicKeys {
		p, err := ssh.ParsePublicKey(key)
		if err != nil {
			log.Print(err)
			break
		}
		pubKeys = append(pubKeys, p)
	}

	return pubKeys, nil
}
