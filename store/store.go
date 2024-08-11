package store

type Store interface {
    GetAll() map[string]string
	Get(key string) (string, bool)
    Set(key, value string)
    Del(key string)
}