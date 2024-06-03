package cmd

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	// "io"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/redis"
)

var ErrKeyNotFound = errors.New("failed to find the key")

type mockDB struct {
	setKeys []string
	delKeys []string
	invalidKeys []string
}

func (m *mockDB) Get(key string) (string, error) {
	if slices.Contains(m.setKeys, key) {
		return "found", nil
	} else {
		return "", ErrKeyNotFound
	}
}

func (m *mockDB) Set(key, val string) {
	m.setKeys = append(m.setKeys, key)
}

func (m *mockDB) Del(key string) string {
	if slices.Contains(m.delKeys, key) {
		return "nil"
	} else if slices.Contains(m.setKeys, key) {
		m.delKeys = append(m.delKeys, key)
		return "1"
	} else {
		return "0"
	}
}

func (m *mockDB) Incr(key string) error {
	if slices.Contains(m.setKeys, key) {
		return nil
	} else if slices.Contains(m.invalidKeys, key) {
		return redis.ErrKeyNotInteger
	} else {
		return nil
	}
}

// type Server struct {
// 	db redis.DbInterface
// 	out io.Writer
// 	// it will contain network related stuff later on
// }

func GetTestServer(db *mockDB) *Server {
	var buf bytes.Buffer
	return &Server{
		db: db,
		out: &buf,
	}
}

// server init
// command comes
// parser parses it and send it to the relevant function like GetAction etc

func TestCommandParser(t *testing.T) {
	t.Run("SET command", func(t *testing.T) {
		input := "SET foo bar"
		expOut := "OK"

		md := &mockDB{}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !slices.Contains(md.setKeys, "foo") {
			t.Errorf("Expected the key to be inside %q but didn't found it", md.setKeys)
		}

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})

	t.Run("SET command with invalid number of arguments (1)", func(t *testing.T) {

	})

	t.Run("SET command with invalid number of arguments (3)", func(t *testing.T) {

	})

	t.Run("GET command with valid key", func(t *testing.T) {
		input := "GET foo"
		expOut := "found"

		md := &mockDB{setKeys: []string{"foo"}}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with invalid key", func(t *testing.T) {
		input := "GET foo"
		expOut := redis.ErrKeyNotFound.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with deleted key", func(t *testing.T) {
		input := "DEL foo"
		expOut := "nil"

		md := &mockDB{delKeys: []string{"foo"}}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})

	t.Run("DEL command with valid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := "1"

		md := &mockDB{setKeys: []string{"foo"}}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !slices.Contains(md.delKeys, "foo") {
			t.Errorf("Expected the key to be inside %q but didn't found it", md.delKeys)
		}

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})

	t.Run("DEL command with invalid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := "0"

		md := &mockDB{}
		s := GetTestServer(md)
		s.ParseCommand(input)

		if !bytes.Contains(s.out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.out.(*bytes.Buffer).String())
		}
	})
}