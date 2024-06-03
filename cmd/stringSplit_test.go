package cmd

import (
	"slices"
	"testing"
	// "errors"
)

func TestStringSplit(t *testing.T) {
	t.Run("command without quotes", func(t *testing.T) {
		input := "GET foo bar"
		expOut := []string{"GET", "foo", "bar"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})
	t.Run("command with quotes", func(t *testing.T) {
		input := `SET foo "bar in quotes"`
		expOut := []string{"SET", "foo", "bar in quotes"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})

	// t.Run("invalid command", func(t *testing.T) {
	// 	input := `SET foo bar"in"quotes"`
	// 	expOut := ErrInvalidCommand

	// 	_, err := StringSplit(input)
	// 	if err == nil {
	// 		t.Fatal("Expected error but got none")
	// 	}

	// 	if errors.Is(err, expOut) {
	// 		return
	// 	}

	// 	t.Fatalf("Unexpected error occured : %v", err)
	// })
}