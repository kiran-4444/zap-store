package zapstore

import (
	"fmt"
	"zap-store/internal/storage"
)

type ZapStore struct {
	StorageEngine storage.StorageEngine
}

// NewZapStore creates a new instance of ZapStore with the provided storage engine
func NewZapStore(engine storage.StorageEngine) *ZapStore {
	return &ZapStore{
		StorageEngine: engine,
	}
}

// Get retrieves a value from the storage engine by key
func (kv *ZapStore) Get(key string) (string, error) {
	return kv.StorageEngine.Get(key)
}

// Set stores a value in the storage engine with the given key
func (kv *ZapStore) Set(key string, value string) error {
	return kv.StorageEngine.Set(key, value)
}

// Delete removes a value from the storage engine by key
func (kv *ZapStore) Delete(key string) error {
	return kv.StorageEngine.Delete(key)
}

var ErrInvalidStorageEngine = fmt.Errorf("invalid storage engine")
