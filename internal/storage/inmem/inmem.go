package inmem

import "fmt"

type InMemStorageEngine struct {
	hashMap map[string]string
}

func (kvs *InMemStorageEngine) Set(key string, value string) error {
	if key == "" {
		return fmt.Errorf("Key cannot be empty")
	}
	kvs.hashMap[key] = value
	return nil
}

func (kvs *InMemStorageEngine) Get(key string) (string, error) {
	if _, ok := kvs.hashMap[key]; !ok {
		return "", fmt.Errorf("Key not found")
	}
	return kvs.hashMap[key], nil
}

func (kvs *InMemStorageEngine) Del(key string) error {
	delete(kvs.hashMap, key)
	return nil
}

func NewInMemStorageEngine() *InMemStorageEngine {
	return &InMemStorageEngine{
		hashMap: make(map[string]string),
	}
}
