package redis

import (
	"errors"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/store"
	// "github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

var ErrKeyNotFound = errors.New("failed to find the key")

type DbInterface interface {
	Set(key, val string)
	Get(key string) (string, error)
	Del(key string) error
}

type Db struct {
	store store.Store
}

func GetNewDB(store store.Store) Db {
	return Db {
		store: store,
	}
}

func(d Db) Set(key, val string) {
	d.store.Set(key, val)
}

func(d Db) Get(key string) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		return "", ErrKeyNotFound
	}

	return val, nil
}

func(d Db) Del(key string) error {
	_, err := d.Get(key)
	if err != nil {
		return err
	}

	d.store.Set(key, "nil")
	return nil
}
// func(d Db) Del(key string) error {
// 	_, ok := d.store.Get(key)
// 	if !ok {
// 		return ErrKeyNotFound
// 	}

// 	d.store.Set(key, "nil")
// 	return nil
// }