package inMemoryStore

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	var mockStore = NewInMemoryStore()

	key := "foo"
	val := "bar"


	t.Run("find store", func(t *testing.T) {
		p, ok := mockStore.Get(key)
		if !ok {
			t.Errorf("Expected true but found %v", ok)
		}

		if p != val {
			t.Fatalf("Expected value to be %q but got %q", val, p)
		}
	})

}