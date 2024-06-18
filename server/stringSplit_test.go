package server

import (
	"slices"
	"testing"
	"errors"
)

func TestStringSplit(t *testing.T) {
	t.Run("command without quotes", func(t *testing.T) {
		input := "SET foo bar"
		expOut := []string{"SET", "foo", "bar"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})

	t.Run("command with value in quotes", func(t *testing.T) {
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

	t.Run("command with key in quotes", func(t *testing.T) {
		input := `SET "foo in quotes" bar`
		expOut := []string{"SET", "foo in quotes", "bar"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})

	t.Run("command with both key and value in quotes", func(t *testing.T) {
		input := `SET "foo in quotes" "bar in quotes"`
		expOut := []string{"SET", "foo in quotes", "bar in quotes"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})

	t.Run("command with everything in quotes", func(t *testing.T) {
		input := `"SET" "foo in quotes" "bar in quotes"`
		expOut := []string{"SET", "foo in quotes", "bar in quotes"}

		out, err := StringSplit(input)
		if err != nil {
			t.Fatalf("Unexpected error occured : %v", err)
		}

		if !slices.Equal(expOut, out) {
			t.Errorf("Expected %q but got %q", expOut, out)
		}
	})

	t.Run("invalid command with quotes in between", func(t *testing.T) {
		input := `SET foo bar"in"quotes"`
		expOut := ErrUnknownCommand

		_, err := StringSplit(input)
		if err == nil {
			t.Fatal("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})

	t.Run("invalid command with unbalanced quotes", func(t *testing.T) {
		input := `SET "foo in quotes "bar in quotes"`
		expOut := ErrUnknownCommand

		_, err := StringSplit(input)
		if err == nil {
			t.Fatal("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})

	t.Run("invalid command with starting in quotes", func(t *testing.T) {
		input := `"SET "foo in quotes "bar in quotes"`
		expOut := ErrUnknownCommand

		_, err := StringSplit(input)
		if err == nil {
			t.Fatal("Expected error but got none")
		}

		if errors.Is(err, expOut) {
			return
		}

		t.Fatalf("Unexpected error occured : %v", err)
	})
}