package lfs

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"sync"
)

type Session struct {
	ID       string
	Username string
}

var (
	SessionStore sync.Map
)

func NewSession(user string) (Session, error) {
	buf := make([]byte, 256)
	_, err := rand.Read(buf)
	if err != nil {
		return Session{}, err
	}
	hash := sha1.Sum(buf)

	sess := Session{ID: fmt.Sprintf("%x", hash), Username: user}
	SessionStore.Store(sess.ID, &sess)
	return sess, nil
}

func FindSession(id string) (Session, error) {
	v, ok := SessionStore.Load(id)
	if ok == false {
		return Session{}, errors.New("session not found")
	}
	sess := v.(*Session)
	return *sess, nil
}
