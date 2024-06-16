package store

type Store interface {
    // GetAll() map[string]string
	Get(key string) (string, bool)
    // Update(key, value string)
    Set(key, value string)
    Del(key string)
}