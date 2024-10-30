package keyvaldb

import (
	"strconv"

	"github.com/justsushant/one2n-go-bootcamp/go-redis/db"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/store"
)

// TODO: leaky abstraction for the store interface; fix the exported struct and make sure that constructor is not available in the entire scope of the project
// TODO: try the composition of interfaces in the struct (look at the ReadWriter interface for inspirtion)

type KeyValueDatabase struct {
	store store.Store
}

func NewDB(store store.Store) *KeyValueDatabase {
	return &KeyValueDatabase{
		store: store,
	}
}

func (d KeyValueDatabase) Set(key, val string) {
	d.store.Set(key, val)
}

func (d KeyValueDatabase) Get(key string) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		return "", db.ErrKeyNotFound
	}

	return val, nil
}

func (d KeyValueDatabase) Del(key string) string {
	_, ok := d.store.Get(key)
	if !ok {
		return db.DELETE_FAILED_MESSAGE
	}

	d.store.Del(key)
	return db.DELETE_SUCCESS_MESSAGE
}

func (d KeyValueDatabase) Incr(key string) (string, error) {
	return d.incrementBy(key, 1)
}

func (d KeyValueDatabase) Incrby(key, incrBy string) (string, error) {
	incrByInt, err := strconv.Atoi(incrBy)
	if err != nil {
		return "", db.ErrKeyNotInteger
	}

	return d.incrementBy(key, incrByInt)
}

func (d KeyValueDatabase) incrementBy(key string, incr int) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, strconv.Itoa(incr))
		return db.INTEGER + " " + strconv.Itoa(incr), nil
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return "", db.ErrKeyNotInteger
	}

	finalValue := valInt + incr
	d.store.Set(key, strconv.Itoa(finalValue))
	return db.INTEGER + " " + strconv.Itoa(finalValue), nil
}

func (d KeyValueDatabase) GetAll() map[string]string {
	return d.store.GetAll()
}
