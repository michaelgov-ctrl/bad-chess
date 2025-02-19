package models

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const MaxSessionAge = 3 * time.Hour

var (
	ErrInvalidCredentials = errors.New("models: invalid credentials")
)

type LazyAuth struct {
	key     string
	userIds map[string]time.Time
	sync.RWMutex
}

func NewLazyAuth() *LazyAuth {
	auth := &LazyAuth{
		key:     "WelcomeToBadChess",
		userIds: make(map[string]time.Time),
	}

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			auth.CleanupUserIds()
		}
	}()

	return auth
}

func (a *LazyAuth) Exists(id string) bool {
	a.RLock()
	defer a.RUnlock()

	if _, ok := a.userIds[id]; !ok {
		return false
	}

	return true
}

func (a *LazyAuth) Authenticate(key string) (string, error) {
	if key != a.key {
		return "", ErrInvalidCredentials
	}

	id := uuid.NewString()

	a.Lock()
	defer a.Unlock()
	a.userIds[id] = time.Now()

	return id, nil
}

func (a *LazyAuth) CleanupUserIds() {
	a.Lock()
	defer a.Unlock()

	for k, v := range a.userIds {
		if time.Since(v) > MaxSessionAge {
			delete(a.userIds, k)
		}
	}
}
