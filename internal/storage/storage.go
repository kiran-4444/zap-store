package storage

type StorageEngine interface {
	Get(string) (string, error)
	Set(string, string) error
	Del(string) error
}
