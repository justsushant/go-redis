package redis

import (
	"errors"
	"testing"
)

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

func GetTestDB() *Db {
	return &Db{
		store: &mockStore{
			vals: make(map[string]string),
		},
	}
}

func TestSet(t *testing.T) {
	key := "foo"
	val := "bar"

	newDB := GetTestDB()

	newDB.Set(key, val)

	if newDB.store.(*mockStore).vals[key] != val {
		t.Errorf("Expected key value pair to be %v/%v but got %v/%v", key, val, newDB.store.(*mockStore).vals[key], val)
	}

	// if newDB.store.vals[key] != val {
	// 	t.Errorf("Expected key value pair to be %v/%v but got %v/%v", key, val, mStore.vals[0], mStore.vals[1])
	// }
}

func TestGet(t *testing.T) {
	t.Run("when key exists", func(t *testing.T) {
		key := "foo"
		val := "bar"

		newDB := GetTestDB()
		newDB.store.Set(key, val)

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


		newDB := GetTestDB()
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
		newDB := GetTestDB()

		key := "foo"
		val := "bar"

		newDB.store.Set(key, val)

		err := newDB.Del(key)

		if err != nil {
			t.Fatalf("error occured: %v", err)
		}

		if val != "bar" {
			t.Errorf("Value doesn't matches, %v bar", val)
		}

		val, _ = newDB.store.Get(key)
		if val != "nil" {
			t.Errorf("Supposed to get nil when getting the key which has been deleted")
		}
	})

	t.Run("when key doesn't exists", func(t *testing.T) {
		newDB := GetTestDB()

		key := "foo"
		err := newDB.Del(key)

		if err == nil {
			t.Fatalf("error didn't occured")
		}

		if errors.Is(err, ErrKeyNotFound) {
			return
		}

		t.Fatalf("unexpected error occured : %v", err)
	})

}
