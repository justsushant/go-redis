package db

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

func (m *mockStore) Del(key string) {
	if m.key == key {
		m.key = ""
		m.val = ""
	}
}

func (m *mockStore) GetAll() map[string]string {
	return map[string]string{
		m.key: m.val,
	}
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
		expOut := val

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		v, err := newDB.Get(key)
		if err != nil {
			t.Fatalf("Unexpected error occured: %v", err)
		}

		if v != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, v)
		}
	})

	t.Run("when key doesn't exists", func(t *testing.T) {
		key := "foo"
		expOut := ErrKeyNotFound

		mockStore := &mockStore{key: "abc", val: "pqr"}
		newDB := &Db{store: mockStore}

		_, err := newDB.Get(key)

		if err == nil {
			t.Fatalf("Error didn't occured: %v", err)
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("unexpected error occured : %v", err)
	})

	// if user performs a get operation on the deleted key, output should be nil
	t.Run("when key is deleted", func(t *testing.T) {
		key := "foo"
		val := "bar"
		expOut := ErrKeyNotFound

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}
		newDB.Del(key)

		_, err := newDB.Get(key)
		if err == nil {
			t.Fatalf("Expected an error but got nil")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})
}

func TestDeleteCommand(t *testing.T) {
	t.Run("when key exists", func(t *testing.T) {
		key := "foo"
		val := "bar"
		expOut := DeleteSuccessMessage

		mockStore := &mockStore{key, val}
		newDB := &Db{store: mockStore}

		out := newDB.Del(key)
		if out != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}

		// if user performs a get operation on the deleted key, output should be nil
		// val, _ = newDB.Get(key)
		// if val != "nil" {
		// 	t.Errorf("Supposed to get nil when getting the key which has been deleted")
		// }
	})

	t.Run("when key doesn't exists", func(t *testing.T) {
		key := "foo"
		expOut := DeleteSuccessMessage

		mockStore := &mockStore{key: "abc", val: "pqr"}
		newDB := &Db{store: mockStore}
		out := newDB.Del(key)

		if out != DeleteFailedMessage {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}
	})
}

func TestIncrCommand(t *testing.T) {
	t.Run("when val is integer", func(t *testing.T) {
		key := "foo"
		val := "4"
		expOut := "(integer) 5"

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		out, err := newDB.Incr(key)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "5" {
			t.Errorf("Supposed to increment value, expecting %d but got %s", 5, mockStore.val)
		}

		if out != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}
	})

	t.Run("when val is not an integer", func(t *testing.T) {
		key := "abc"
		val := "pqr"
		expOut := ErrKeyNotInteger

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		_, err := newDB.Incr(key)
		if err == nil {
			t.Fatalf("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})

	t.Run("when val doesn't exist", func(t *testing.T) {
		key := "foo"
		expOut := "(integer) 1"
		
		mockStore := &mockStore{}
		newDB := &Db{store: mockStore}

		out, err := newDB.Incr(key)
		if err != nil {
			t.Fatalf("Unexpected error occured: %v", err)
		}

		if mockStore.val != DefaultIntegerValue {
			t.Fatalf("Expected %s but got %s", DefaultIntegerValue, mockStore.val)
		}

		if out != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}
	})
}

func TestIncrByCommand(t *testing.T) {
	t.Run("when val is integer", func(t *testing.T) {
		key := "foo"
		val := "5"
		incrVal := "20"
		expOut := "(integer) 25"

		mockStore := &mockStore{key, val}
		newDB := &Db{store: mockStore}

		out, err := newDB.Incrby(key, incrVal)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "25" {
			t.Errorf("Supposed to increment to val, expecting %d but got %s", 25, mockStore.val)
		}

		if out != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}
	})

	t.Run("when val is not an integer", func(t *testing.T) {
		key := "foo"
		val := "bar"
		expOut := ErrKeyNotInteger

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		_, err := newDB.Incrby(key, "10")
		if err == nil {
			t.Fatalf("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})

	t.Run("when passed val is not an integer", func(t *testing.T) {
		key := "foo"
		val := "bar"
		expOut := ErrKeyNotInteger

		mockStore := &mockStore{key: key, val: val}
		newDB := &Db{store: mockStore}

		_, err := newDB.Incrby(key, "bar+")
		if err == nil {
			t.Fatalf("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})

	t.Run("when val doesn't exist", func(t *testing.T) {
		key := "foo"
		val := "28"
		expOut := "(integer) 28"

		mockStore := &mockStore{}
		newDB := &Db{store: mockStore}

		out, err := newDB.Incrby(key, val)
		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if mockStore.val != "28" {
			t.Errorf("Supposed to set the key, expecting %d but got %s", 23, mockStore.val)
		}

		if out != expOut {
			t.Errorf("Expected the value to be %v instead of %v", expOut, out)
		}
	})
}
