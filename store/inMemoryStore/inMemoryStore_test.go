package inMemoryStore

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	key := "foo"
	val := "bar"

	t.Run("set op", func(t *testing.T) {
		var dummyStore = NewInMemoryStore()
		dummyStore.Set(key, val)

		v, ok := dummyStore.data[key]
		if !ok {
			t.Fatalf("Failed to set the key %s in store", key)
		}
		if v != val {
			t.Errorf("Expected the value %s for the key %s but found %s", val, key, v)
		}
	})

	t.Run("get op for existing key", func(t *testing.T) {
		var dummyStore = NewInMemoryStore()
		dummyStore.data[key] = val

		v, ok := dummyStore.Get(key)
		if !ok {
			t.Fatalf("Expected the value %s for key %s but didn't got any", val, key)
		}
		if v != val {
			t.Errorf("Expected the value %s for the key %s but found %s", val, key, v)
		}
	})

	t.Run("get op for non-existent key", func(t *testing.T) {
		var dummyStore = NewInMemoryStore()

		v, ok := dummyStore.Get(key)
		if ok {
			t.Fatalf("Didn't expected to find the value for key %s but got %s", key, v)
		}
	})

	t.Run("del op", func(t *testing.T) {
		var dummyStore = NewInMemoryStore()
		dummyStore.data[key] = val

		dummyStore.Del(key)
		v, ok := dummyStore.data[key]
		if ok {
			t.Fatalf("Didn't expected to find the value for key %s but got %s", key, v)
		}
	})

	t.Run("get all op", func(t *testing.T) {
		var dummyStore = NewInMemoryStore()
		dummyStore.data["key1"] = "val1"
		dummyStore.data["key2"] = "val2"
		dummyStore.data["key3"] = "val3"

		result := dummyStore.GetAll()

		v1, ok := result["key1"]
		if !ok {
			t.Fatalf("Expected to find the key %s with value %s but didn't got any", "key1", "val1")
		}
		if v1 != "val1" {
			t.Errorf("Expected the value %s for the key %s but found %s", val, key, v1)
		}

		v2, ok := result["key1"]
		if !ok {
			t.Fatalf("Expected to find the key %s with value %s but didn't got any", "key1", "val1")
		}
		if v2 != "val1" {
			t.Errorf("Expected the value %s for the key %s but found %s", val, key, v1)
		}

		v3, ok := result["key1"]
		if !ok {
			t.Fatalf("Expected to find the key %s with value %s but didn't got any", "key1", "val1")
		}
		if v3 != "val1" {
			t.Errorf("Expected the value %s for the key %s but found %s", val, key, v1)
		}
	})

}