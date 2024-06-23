package server

import (
	"slices"
	"testing"
	"errors"
)

func TestStringSplit(t *testing.T) {
	md := &mockDB{}
	s := GetTestServer(md, nil)

	testCases := []struct{
		name string
		input string
		expOut []string
		isError bool
		err error
	}{
		{"command without quotes", "SET foo bar", []string{"SET", "foo", "bar"}, false, nil},
		{"command with value in quotes", "SET foo \"bar in quotes\"", []string{"SET", "foo", "bar in quotes"}, false, nil},
		{"command with key in quotes", "SET \"foo in quotes\" bar", []string{"SET", "foo in quotes", "bar"}, false, nil},
		{"command with both key and value in quotes", "SET \"foo in quotes\" \"bar in quotes\"", []string{"SET", "foo in quotes", "bar in quotes"}, false, nil},
		{"command with everything in quotes", "\"SET\" \"foo in quotes\" \"bar in quotes\"", []string{"SET", "foo in quotes", "bar in quotes"}, false, nil},
		{"invalid command with quotes in between", "SET foo bar\"in\"quotes", nil, true, ErrUnknownCommand},
		{"invalid command with unbalanced quotes", "SET \"foo in quotes \"bar in quotes\"", nil, true, ErrUnknownCommand},
		{"invalid command with starting in quotes", "\"SET \"foo in quotes \"bar in quotes\"", nil, true, ErrUnknownCommand},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := s.stringSplit(tc.input)

			if tc.isError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				if errors.Is(err, tc.err) {
					return
				}

				t.Fatalf("Unexpected error occured : %v", err)
			}

			if !slices.Equal(tc.expOut, out) {
				t.Errorf("Expected %q but got %q", tc.expOut, out)
			}
		})
	}
}