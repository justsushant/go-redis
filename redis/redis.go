package redis

import (
	"errors"
	"strconv"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/store"
	// "github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

var ErrKeyNotFound = errors.New("(nil)")
// var ErrKeyNotFound = errors.New("failed to find the key")
var ErrKeyNotInteger = errors.New("value is not an integer or out of range")
var SetSuccessMessage = "OK"
var DeleteSuccessMessage = "(integer) 1"
var DeleteFailedMessage = "(integer) 0"
var DefaultIntegerValue = "1"
var Integer = "(integer)"

type DbInterface interface {
	Set(key, val string)
	Get(key string) (string, error)
	Del(key string) string
	Incr(key string) (string, error)
	Incrby(key, val string) (string, error)
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

	d.store.Del(key)
	return DeleteSuccessMessage
}

func(d Db) Incr(key string) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, DefaultIntegerValue)
		return SetSuccessMessage, nil
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return "", ErrKeyNotInteger
	}

	incrVal := i+1
	d.store.Set(key, strconv.Itoa(incrVal))
	return Integer + " " + strconv.Itoa(incrVal), nil
}

func(d Db) Incrby(key, i string) (string, error) {
	num, err := strconv.Atoi(i)
	if err != nil {
		return "", ErrKeyNotInteger
	}

	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, strconv.Itoa(num))
		return Integer + " " + i, nil
	}

	vali, err := strconv.Atoi(val)
	if err != nil {
		return "", ErrKeyNotInteger
	}
	
	incrVal := num+vali
	d.store.Set(key, strconv.Itoa(incrVal))
	return Integer + " " + strconv.Itoa(incrVal), nil
}