package inMemoryStore

import (
	"testing"

	// store "github.com/justsushant/one2n-go-bootcamp/redis-go/store"
)

func TestInMemoryStore(t *testing.T) {
	var mockStore = NewInMemoryStore()

	key := "foo"
	val := "bar"

	// t.Run("get empty store", func(t *testing.T) {
	// 	m := mockStore.GetAll()

	// 	if len(m) > 0 {
	// 		t.Fatalf("Expected zero key-val pairs but found %d pairs", len(m))
	// 	}
	// })

	// t.Run("update store", func(t *testing.T) {
	// 	mockStore.Update(key, val)

	// 	if len(mockStore.data) != 1 {
	// 		t.Fatalf("Expected 1 key-value pair but found %d pairs", len(mockStore.data))
	// 	}
	// })

	t.Run("find store", func(t *testing.T) {
		p, ok := mockStore.Get(key)
		if !ok {
			t.Errorf("Expected true but found %v", ok)
		}

		if p != val {
			t.Fatalf("Expected value to be %q but got %q", val, p)
		}
	})

	// t.Run("get store", func(t *testing.T) {
	// 	m := mockStore.GetAll()

	// 	if len(m) != 1 {
	// 		t.Fatalf("Expected 1 key-val pairs but found %d pairs", len(m))
	// 	}
	// })
}