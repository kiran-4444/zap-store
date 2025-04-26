package storage

type StorageEngine interface {
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
	Close() error
}
