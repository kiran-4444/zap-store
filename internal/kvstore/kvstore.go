package kvstore

import (
	"fmt"
)

type KVStore struct {
	hashMap map[string]string
}

func (kvs *KVStore) Set(key string, value string) error {
	if key == "" {
		return fmt.Errorf("Key cannot be empty")
	}

	kvs.hashMap[key] = value
	return nil
}

func (kvs *KVStore) Get(key string) (string, error) {
	if _, ok := kvs.hashMap[key]; !ok {
		return "", fmt.Errorf("Key not found")
	}

	return kvs.hashMap[key], nil
}

func (kvs *KVStore) Del(key string) {
	delete(kvs.hashMap, key)
}

func NewKVStore() *KVStore {
	return &KVStore{
		hashMap: make(map[string]string),
	}
}
