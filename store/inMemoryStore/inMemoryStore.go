package inMemoryStore

import (
	"sync"
)

type InMemoryStore struct {
	data map[string]string
	sync.RWMutex
}

func (i *InMemoryStore) GetAll() map[string]string {
	i.RLock()
	defer i.RUnlock()
	return i.data
}

func (i *InMemoryStore) Set(key, value string) {
	i.Lock()
	defer i.Unlock()
	i.data[key] = value
}

func (i *InMemoryStore) Get(key string) (string, bool) {
	i.RLock()
	defer i.RUnlock()
	proxy, ok := i.data[key]
    return proxy, ok
}

func (i *InMemoryStore) Del(key string) {
	i.RLock()
	defer i.RUnlock()
	delete(i.data, key)
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}
