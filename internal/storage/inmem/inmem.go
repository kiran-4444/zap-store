package inmem

import (
	"fmt"
	"sync"
)

type InMemStorageEngine struct {
	hashMap map[string]string
	lock    sync.Mutex
}

func (kvs *InMemStorageEngine) Set(key string, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	kvs.lock.Lock()
	defer kvs.lock.Unlock()

	kvs.hashMap[key] = value
	return nil
}

func (kvs *InMemStorageEngine) Get(key string) (string, error) {
	kvs.lock.Lock()
	defer kvs.lock.Unlock()

	if _, ok := kvs.hashMap[key]; !ok {
		return "", fmt.Errorf("key not found")
	}
	return kvs.hashMap[key], nil
}

func (kvs *InMemStorageEngine) Del(key string) error {
	kvs.lock.Lock()
	defer kvs.lock.Unlock()

	delete(kvs.hashMap, key)
	return nil
}

func (kvs *InMemStorageEngine) Close() error {
	return nil
}

func NewInMemStorageEngine() *InMemStorageEngine {
	return &InMemStorageEngine{
		hashMap: make(map[string]string),
	}
}
