package db

type Database interface {
	GetAll() map[string]string
	Set(key, val string)
	Get(key string) (string, error)
	Del(key string) string
	Incr(key string) (string, error)
	Incrby(key, val string) (string, error)
}
