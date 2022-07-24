package core

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	_keyspaceTemplate = ":%s:%s:%s:"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type Data struct {
	Value string
	Time  time.Time
}

type ConfigStore interface {
	Set(ns, typ, name string, data Data) error
	Get(ns, typ, name string) (*Data, error)
}

func NewConfigStore() ConfigStore {
	return &configStore{cache.New(5*time.Minute, 10*time.Minute)}
}

type configStore struct {
	store *cache.Cache
}

func (s *configStore) Set(ns, typ, name string, data Data) error {
	s.store.Set(fmt.Sprintf(_keyspaceTemplate, ns, typ, name), data, cache.NoExpiration)
	return nil
}

func (s *configStore) Get(ns, typ, name string) (*Data, error) {
	rawData, ok := s.store.Get(fmt.Sprintf(_keyspaceTemplate, ns, typ, name))
	if !ok {
		return nil, ErrNotFound
	}
	data := rawData.(Data)
	return &data, nil
}
