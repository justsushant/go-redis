package cmd

import (
	"testing"
	"errors"

	// "github.com/justsushant/one2n-go/-bootcamp/redis-go/redis"
)

var ErrKeyNotFound = errors.New("failed to find the key")

type MockDB struct {
	store map[string]string
}

func (m *MockDB) Get(key string) (string, bool) {
	val, ok := m.store[key]
	if !ok {
		return "", false
	}

	return val, true
}

func (m *MockDB) Set(key, value string) {
	m.store[key] = value
}

func (m *MockDB) Del(key string) error {
	_, ok := m.Get(key)
	if ok != true {
		return ErrKeyNotFound
	}

	m.store.Set(key, "nil")
	return nil
}

func GetTestDB() *MockDB {
	return &MockDB{
		store: make(map[string]string),
	}
}


func TestSetAction(t *testing.T) {
	// key := "foo"
	// val := "bar"

	// db := GetTestDB()

	// SetAction(db, key, val)
}