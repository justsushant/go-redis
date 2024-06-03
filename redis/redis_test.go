package redis

import (
	"errors"
	"testing"
)

type mockStore struct {
	key string
	val string
}

func (m *mockStore) Get(key string) (string, bool) {
	if m.key == key {
		return m.val, true
	}
	return "", false
}

func (m *mockStore) Set(key, val string) {
	m.key = key
	m.val = val
}

func GetTestDB(key, val string) *Db {
	return &Db{
		store: &mockStore{
			key: key,
			val: val,
		},
	}
}

func TestSet(t *testing.T) {
	key := "foo"
	val := "bar"

	mockStore := &mockStore{key: key, val: val}
	newDB := &Db{store: mockStore}

	newDB.Set(key, val)

	if mockStore.key != key || mockStore.val != val {
		t.Errorf("Expected key value pair to be %v/%v but got %v/%v", key, val, mockStore.key, mockStore.val)
	}
}

func TestGet(t *testing.T) {
	t.Run("when key exists", func(t *testing.T) {
		key := "foo"
		val := "bar"

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		v, err := newDB.Get(key)
		if err != nil {
			t.Fatalf("Unexpected error occured: %v", err)
		}

		if v != val {
			t.Errorf("Expected the value to be %v instead of %v", val, v)
		}
	})

	t.Run("when key doesn't exists", func(t *testing.T) {
		key := "foo"

		mockStore := &mockStore{key: "abc", val: "pqr"}
		newDB := &Db{store: mockStore}

		_, err := newDB.Get(key)

		if err == nil {
			t.Fatalf("Error didn't occured: %v", err)
		}

		if errors.Is(err, ErrKeyNotFound) {
			return
		}

		t.Fatalf("unexpected error occured : %v", err)
	})
}

func TestDeleteCommand(t *testing.T) {
	t.Run("when key exists", func(t *testing.T) {
		key := "foo"
		val := "bar"

		mockStore := &mockStore{key, val}
		newDB := &Db{store: mockStore}

		out := newDB.Del(key)
		if out != DeleteSuccessMessage {
			t.Fatalf("Didnot found the success message, got %q instead", out)
		}

		// if user performs a get operation on the deleted key, output should be nil
		val, _ = newDB.Get(key)
		if val != "nil" {
			t.Errorf("Supposed to get nil when getting the key which has been deleted")
		}
	})

	t.Run("when key doesn't exists", func(t *testing.T) {
		mockStore := &mockStore{key: "abc", val: "pqr"}
		newDB := &Db{store: mockStore}

		key := "foo"
		out := newDB.Del(key)

		if out != DeleteFailedMessage {
			t.Fatalf("Didnot found the failed message, got %q instead", out)
		}
	})
}

func TestIncrCommand(t *testing.T) {
	t.Run("when val is integer", func(t *testing.T) {
		key := "foo"
		val := "5"

		mockStore := &mockStore{key, val}
		newDB := &Db{store: mockStore}

		err := newDB.Incr(key)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "6" {
			t.Errorf("Supposed to increment to val, expecting %d but got %s", 6, mockStore.val)
		}
	})

	t.Run("when val is not an integer", func(t *testing.T) {
		key := "abc"
		val := "pqr"
		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		err := newDB.Incr(key)
		if err == nil {
			t.Fatalf("Expected error but got none")
		}

		if errors.Is(err, ErrKeyNotInteger) {
			return
		}

		t.Fatalf("unexpected error occured : %v", err)
	})

	t.Run("when val doesn't exist", func(t *testing.T) {
		key := "abc"
		mockStore := &mockStore{}
		newDB := &Db{store: mockStore}

		err := newDB.Incr(key)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != DefaultIntegerValue {
			t.Fatalf("Expected %s but got %s", DefaultIntegerValue, mockStore.val)
		}
	})
}

func TestIncrByCommand(t *testing.T) {
	t.Run("when val is integer", func(t *testing.T) {
		key := "foo"
		val := "5"

		mockStore := &mockStore{key, val}
		newDB := &Db{store: mockStore}

		err := newDB.IncrBy(key, 20)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "25" {
			t.Errorf("Supposed to increment to val, expecting %d but got %s", 25, mockStore.val)
		}
	})

	t.Run("when val is not an integer", func(t *testing.T) {
		key := "abc"
		val := "pqr"
		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		err := newDB.IncrBy(key, 8)
		if err == nil {
			t.Fatalf("Expected error but got none")
		}

		if errors.Is(err, ErrKeyNotInteger) {
			return
		}

		t.Fatalf("unexpected error occured : %v", err)
	})

	t.Run("when val doesn't exist", func(t *testing.T) {
		key := "abc"
		val := 23
		mockStore := &mockStore{}
		newDB := &Db{store: mockStore}

		err := newDB.IncrBy(key, val)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "23" {
			t.Errorf("Supposed to increment to val, expecting %d but got %s", 23, mockStore.val)
		}
	})
}
