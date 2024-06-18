package server

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
)

var ErrKeyNotFound = errors.New("failed to find the key")

type mockDB struct {
	key string
	val string
}

func (m *mockDB) Get(key string) (string, error) {
	if m.key == key {
		return m.val, nil
	} else {
		return "", ErrKeyNotFound
	}
}

func (m *mockDB) Set(key, val string) {
	m.key = key
	m.val = val
}

func (m *mockDB) Del(key string) string {
	if m.key == key {
		m.key = ""
		m.val = ""
		return "(integer) 1"
	} else {
		return "(integer) 0"
	}
}

func (m *mockDB) Incr(key string) (string, error) {
	if m.key == key {
		i, err := strconv.Atoi(m.val)
		if err != nil {
			return "", db.ErrKeyNotInteger
		}
		i+=1
		m.val = strconv.Itoa(i)
		return "(integer) " + strconv.Itoa(i), nil
	} else {
		m.key = key
		m.val = "1"
		return MssgOK, nil
	}
}

func (m *mockDB) Incrby(key, num string) (string, error) {
	i2, err := strconv.Atoi(num)
	if err != nil {
		return "", db.ErrKeyNotInteger
	}

	if m.key == key {
		i, err := strconv.Atoi(m.val)
		if err != nil {
			return "", db.ErrKeyNotInteger
		}
		incrByVal := i+i2
		m.val += strconv.Itoa(incrByVal)
		return "(integer) " + strconv.Itoa(incrByVal), nil
	} else {
		m.key = key
		m.val = num
		return "(integer) " + num, nil
	}
}

func (m *mockDB) GetAll() map[string]string {
	if m.key == "one" {
		return map[string]string{"foo": "bar"}
	} else if m.key == "multiple" {
		return map[string]string{"foo": "bar", "counter": "13"}
	} else {
		return map[string]string{}
	}
}


func GetTestServer(db *mockDB) *Server {
	var buf bytes.Buffer
	return &Server{
		Db: db,
		Out: &buf,
	}
}

func TestCommandParser(t *testing.T) {
	t.Run("SET command", func(t *testing.T) {
		input := "SET foo bar"
		expOut := MssgOK

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if md.key != "foo" {
			t.Errorf("Expected the key to be %q but didn't found it", md.key)
		}

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("SET command with invalid number of arguments (1)", func(t *testing.T) {
		input := "SET foo"
		expOut := ErrWrongNumberOfArgs.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("SET command with invalid number of arguments (3)", func(t *testing.T) {
		input := "SET foo bar extra"
		expOut := ErrWrongNumberOfArgs.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with valid key", func(t *testing.T) {
		input := "GET foo"
		expOut := strconv.Quote("bar")

		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with deleted key", func(t *testing.T) {
		input := "GET foo"
		expOut := db.ErrKeyNotFound.Error()

		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md)
		s.handleCommand("DEL foo")
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with invalid key", func(t *testing.T) {
		input := "GET foo"
		expOut := db.ErrKeyNotFound.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("GET command with invalid number of args (2)", func(t *testing.T) {
		input := "GET foo bar"
		expOut := ErrWrongNumberOfArgs.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("DEL command with valid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := db.DeleteSuccessMessage

		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if md.key != "" {
			t.Errorf("Expected the key to be deleted but found %q", md.key)
		}

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("DEL command with invalid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := db.DeleteFailedMessage

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCR command with valid key", func(t *testing.T) {
		input := "INCR foo"
		expOut := "(integer) 5"

		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCR command with invalid key", func(t *testing.T) {
		input := "INCR foo"
		expOut := db.SetSuccessMessage

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCR command with invalid value (string)", func(t *testing.T) {
		input := "INCR foo"
		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCR command with invalid value (float)", func(t *testing.T) {
		input := "INCR foo"
		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{key: "foo", val: "10.5"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with valid key", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := "(integer) 9"

		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid key", func(t *testing.T) {
		input := "INCRBY foo 8"
		expOut := "(integer) 8"

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid key and passed val is string", func(t *testing.T) {
		input := "INCRBY foo bar"
		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})
	t.Run("INCRBY command with passed key is string with integer val set", func(t *testing.T) {
		input := "INCRBY foo bar"

		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid value (string)", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid value (float)", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := db.ErrKeyNotInteger.Error()

		md := &mockDB{key: "foo", val: "10.5"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid number of arguments (1)", func(t *testing.T) {
		input := "INCRBY foo"
		expOut := ErrWrongNumberOfArgs.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("INCRBY command with invalid number of arguments (3)", func(t *testing.T) {
		input := "INCRBY foo bar extra"
		expOut := ErrWrongNumberOfArgs.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	
	t.Run("COMPACT command with one key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut := "SET foo bar\n"

		md := &mockDB{key: "one"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("COMPACT command with two key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut1 := "SET foo bar\n"
		expOut2 := "SET counter 13\n"

		md := &mockDB{key: "multiple"}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut1)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut1, s.Out.(*bytes.Buffer).String())
		}
		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut2)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut2, s.Out.(*bytes.Buffer).String())
		}
	})

	t.Run("COMPACT command with no key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut := "(nil)\n"

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})


	t.Run("Invalid command", func(t *testing.T) {
		input := "gibberish foo bar"
		expOut := ErrUnknownCommand.Error()

		md := &mockDB{}
		s := GetTestServer(md)
		s.handleCommand(input)

		if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, s.Out.(*bytes.Buffer).String())
		}
	})
}

func TestCommandParserWithMulti(t *testing.T) {
	t.Run("MULTI command with EXEC", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "GET foo", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", "QUEUED", "1) OK\n2) \"bar\""}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI command with DISCARD", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "GET foo", "DISCARD"}
		expOut := []string{MssgOK, "QUEUED", "QUEUED", MssgOK}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI command with EXEC with previous errors (invalid arguments)", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "INCRBY foo 5 6", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrWrongNumberOfArgs.Error(), ErrTranAbortedDueToPrevError.Error()}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI command with EXEC with previous errors (invalid command)", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "RANDOM NONSENSE", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrUnknownCommand.Error(), ErrTranAbortedDueToPrevError.Error()}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("EXEC command without MULTI", func(t *testing.T) {
		inputArr := []string{"EXEC"}
		expOut := []string{ErrExecWithoutMulti.Error()}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("DISCARD command without MULTI", func(t *testing.T) {
		inputArr := []string{"DISCARD"}
		expOut := []string{ErrDiscardWithoutMulti.Error()}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI calls nested", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "MULTI", "INCR foo", "GET foo", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrMultiCommandNested.Error(), "QUEUED", "QUEUED", "1) OK\n2) (integer) 6\n3) \"6\"\n"}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI & EXEC without any commands", func(t *testing.T) {
		inputArr := []string{"MULTI", "EXEC"}
		expOut := []string{MssgOK, MssgEmptyArray}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})

	t.Run("MULTI command with argument to EXEC", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "EXEC GVK", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrWrongNumberOfArgs.Error(), ErrExecWithoutMulti.Error()}
		md := &mockDB{}
		s := GetTestServer(md)

		for i, input := range inputArr {
			s.handleCommand(input)

			
			if !bytes.Contains(s.Out.(*bytes.Buffer).Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], s.Out.(*bytes.Buffer).String())
			}

			s.Out.(*bytes.Buffer).Reset()
		}
	})
}
