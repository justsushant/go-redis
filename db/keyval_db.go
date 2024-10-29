package db

import (
	"strconv"

	"github.com/justsushant/one2n-go-bootcamp/go-redis/store"
)

// TODO: Put the errors & constants in their own file
// TODO: Put the interface away from their implementation (like store package)
// TODO: leaky abstraction for the store interface; fix the exported struct and make sure that constructor is not available in the entire scope of the project
// TODO: try the composition of interfaces in the struct (look at the ReadWriter interface for inspirtion)
// TODO: declare the constants  with const and name them according to conventions (look it up)
// TODO: extract the logic into one unexposed function and call them from actual exposed funcs after proper validation

type KeyValueDatabase struct {
	store store.Store
}

func GetNewDB(store store.Store) KeyValueDatabase {
	return KeyValueDatabase{
		store: store,
	}
}

func (d KeyValueDatabase) Set(key, val string) {
	d.store.Set(key, val)
}

func (d KeyValueDatabase) Get(key string) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		return "", ErrKeyNotFound
	}

	return val, nil
}

func (d KeyValueDatabase) Del(key string) string {
	_, ok := d.store.Get(key)
	if !ok {
		return DELETE_FAILED_MESSAGE
	}

	d.store.Del(key)
	return DELETE_SUCCESS_MESSAGE
}

func (d KeyValueDatabase) Incr(key string) (string, error) {
	return d.incrementBy(key, 1)
}

func (d KeyValueDatabase) Incrby(key, incrBy string) (string, error) {
	incrByInt, err := strconv.Atoi(incrBy)
	if err != nil {
		return "", ErrKeyNotInteger
	}

	return d.incrementBy(key, incrByInt)
}

func (d KeyValueDatabase) incrementBy(key string, incr int) (string, error) {
	val, ok := d.store.Get(key)
	if !ok {
		d.store.Set(key, strconv.Itoa(incr))
		return INTEGER + " " + strconv.Itoa(incr), nil
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return "", ErrKeyNotInteger
	}

	finalValue := valInt + incr
	d.store.Set(key, strconv.Itoa(finalValue))
	return INTEGER + " " + strconv.Itoa(finalValue), nil
}

func (d KeyValueDatabase) GetAll() map[string]string {
	return d.store.GetAll()
}
