package main

import (
    "errors"
    "testing"
)

type mockStore struct {
    vals map[string]string
}

func (m *mockStore) Get(key string) (string, bool) {
    val, ok := m.vals[key]
    return val, ok
}

func (m *mockStore) Set(key, value string) {
    m.vals[key] = value
}

func (m *mockStore) Del(key string) {
    delete(m.vals, key)
}

func GetTestDB() *Db {
    return &Db{
        store: &mockStore{
            vals: make(map[string]string),
        },
    }
}

// ... rest of your tests ...

func TestGetNonExistentKey(t *testing.T) {
    key := "foo"

    mStore := &mockStore{vals: make(map[string]string)}
    newDB := GetNewDB(mStore)

    _, err := newDB.Get(key)
    if !errors.Is(err, ErrKeyNotFound) {
        t.Fatalf("Expected error to be %v, got %v", ErrKeyNotFound, err)
    }
}

func TestDeleteCommand(t *testing.T) {
    t.Run("when key exists", func(t *testing.T) {
        mStore := &mockStore{vals: make(map[string]string)}
        newDB := GetNewDB(mStore)
        
        key := "foo"
        val := "bar"

        newDB.store.Set(key, val)

        err := newDB.Del(key)

        if err != nil {
            t.Fatalf("error occured: %v", err)
        }

        _, ok := newDB.store.Get(key)
        if ok {
            t.Errorf("Key was not deleted")
        }
    })

    // ... rest of your tests ...
}


type mockStore struct {
    vals map[string]string
}

func (m *mockStore) Get(key string) (string, bool) {
    val, ok := m.vals[key]
    return val, ok
}

func (m *mockStore) Set(key, value string) {
    m.vals[key] = value
}

func (m *mockStore) Del(key string) {
    delete(m.vals, key)
}


type MockDB struct {
	store mockStore
}

type mockStore struct {
	vals map[string]string
}

func (m *mockStore) Get(key string) (string, bool) {
	val, ok := m.vals[key]
	if !ok {
		return "", false
	}

	return val, true
}

func (m *mockStore) Set(key, value string) {
	m.vals[key] = value
}