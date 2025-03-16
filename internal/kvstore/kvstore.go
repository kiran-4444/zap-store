package kvstore

import (
	"fmt"
	"kv-store/internal/storage"
)

type KVStore struct {
	StorageEngine storage.StorageEngine
}

// NewKVStore creates a new instance of KVStore with the provided storage engine
func NewKVStore(engine storage.StorageEngine) *KVStore {
	return &KVStore{
		StorageEngine: engine,
	}
}

// Get retrieves a value from the storage engine by key
func (kv *KVStore) Get(key string) (string, error) {
	return kv.StorageEngine.Get(key)
}

// Set stores a value in the storage engine with the given key
func (kv *KVStore) Set(key string, value string) error {
	return kv.StorageEngine.Set(key, value)
}

// Delete removes a value from the storage engine by key
func (kv *KVStore) Del(key string) error {
	return kv.StorageEngine.Del(key)
}

var ErrInvalidStorageEngine = fmt.Errorf("invalid storage engine")
