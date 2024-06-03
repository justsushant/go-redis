package redis

import (
	"errors"
	"strconv"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/store"
	// "github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

var ErrKeyNotFound = errors.New("failed to find the key")
var ErrKeyNotInteger = errors.New("requested key is not integer")
var DeleteSuccessMessage = "1"
var DeleteFailedMessage = "0"
var DefaultIntegerValue = "1"

type DbInterface interface {
	Set(key, val string)
	Get(key string) (string, error)
	Del(key string) string
	// Incr(key string) error
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

func(d Db) Del(key string) string {
	_, ok := d.store.Get(key)
	if !ok {
		return DeleteFailedMessage
	}

	d.store.Set(key, "nil")
	return DeleteSuccessMessage
}

func(d Db) Incr(key string) error {
	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, DefaultIntegerValue)
		return nil
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return ErrKeyNotInteger
	}

	d.store.Set(key, strconv.Itoa(i+1))
	return nil
}

func(d Db) IncrBy(key string, i int) error {
	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, strconv.Itoa(i))
		return nil
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return ErrKeyNotInteger
	}

	d.store.Set(key, strconv.Itoa(num+i))
	return nil
}