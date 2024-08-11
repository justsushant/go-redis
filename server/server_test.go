package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"testing"

	"github.com/justsushant/one2n-go-bootcamp/go-redis/db"
	"google.golang.org/grpc/test/bufconn"
)

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
		i += 1
		m.val = strconv.Itoa(i)
		return "(integer) " + strconv.Itoa(i), nil
	} else {
		m.key = key
		m.val = "1"
		return "(integer) 1", nil
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
		incrByVal := i + i2
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

func GetTestServer(md *mockDB, ln net.Listener) *Server {
	return &Server{
		Db:       map[int]db.DbInterface{0: md},
		Listener: ln,
	}
}

func TestCommandParser(t *testing.T) {
	testCases := []struct {
		name   string
		key    string
		val    string
		input  string
		expOut string
	}{
		{"SET command", "", "", "SET foo bar", MssgOK},
		{"SET command with invalid number of arguments (1)", "", "", "SET foo", ErrWrongNumberOfArgs.Error()},
		{"SET command with invalid number of arguments (3)", "", "", "SET foo bar extra", ErrWrongNumberOfArgs.Error()},
		{"GET command with valid key", "foo", "bar", "GET foo", strconv.Quote("bar")},
		{"GET command with invalid key", "", "", "GET foo", db.ErrKeyNotFound.Error()},
		{"GET command with invalid number of args (2)", "", "", "GET foo bar", ErrWrongNumberOfArgs.Error()},
		{"DEL command with valid key", "foo", "bar", "DEL foo", db.DeleteSuccessMessage},
		{"DEL command with invalid key", "", "", "DEL foo", db.DeleteFailedMessage},
		{"DEL command with invalid number of args (2)", "", "", "DEL foo bar", ErrWrongNumberOfArgs.Error()},
		{"INCR command with valid key", "foo", "4", "INCR foo", "(integer) 5"},
		{"INCR command with invalid key", "", "", "INCR foo", "(integer) 1"},
		{"INCR command with invalid number of args (2)", "", "", "INCR foo bar", ErrWrongNumberOfArgs.Error()},
		{"INCR command with invalid value (string)", "foo", "bar", "INCR foo", db.ErrKeyNotInteger.Error()},
		{"INCR command with invalid value (float)", "foo", "10.5", "INCR foo", db.ErrKeyNotInteger.Error()},
		{"INCRBY command with valid key", "foo", "4", "INCRBY foo 5", "(integer) 9"},
		{"INCRBY command with invalid key", "", "", "INCRBY foo 8", "(integer) 8"},
		{"INCRBY command with invalid key and passed val is string", "", "", "INCRBY foo bar", db.ErrKeyNotInteger.Error()},
		{"INCRBY command with valid key and passed val is string", "foo", "4", "INCRBY foo bar", db.ErrKeyNotInteger.Error()},
		{"INCRBY command with invalid value (string)", "foo", "bar", "INCRBY foo 5", db.ErrKeyNotInteger.Error()},
		{"INCRBY command with invalid value (float)", "foo", "10.5", "INCRBY foo 5", db.ErrKeyNotInteger.Error()},
		{"INCRBY command with invalid number of arguments (1)", "foo", "4", "INCRBY foo", ErrWrongNumberOfArgs.Error()},
		{"INCRBY command with invalid number of arguments (3)", "foo", "4", "INCRBY foo 5 extra", ErrWrongNumberOfArgs.Error()},
		{"COMPACT command with one key-val pair", "one", "1", "COMPACT", "SET foo bar\n"},
		// {"COMPACT command with two key-val pair", "multiple", "2", "COMPACT", "SET foo bar\nSET counter 13\n"},
		{"COMPACT command with no key-val pair", "", "", "COMPACT", "(nil)\n"},
		{"SELECT command with invalid number of arguments (2)", "", "", "SELECT 1 4", ErrWrongNumberOfArgs.Error()},
		{"SELECT command with invalid type of argument (string)", "", "", "SELECT foo", db.ErrKeyNotInteger.Error()},
		{"SELECT command with invalid range of argument (not in 0-15)", "", "", "SELECT 24", ErrDBIndexOutOfRange.Error()},
		{"Invalid command", "", "", "gibberish foo bar", ErrUnknownCommand.Error()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input
			expOut := tc.expOut

			var buf bytes.Buffer
			md := &mockDB{}
			if tc.key != "" {
				md.key = tc.key
			}
			if tc.val != "" {
				md.val = tc.val
			}

			s := GetTestServer(md, nil)
			s.handleCommand(input, &buf, &ConnContext{})

			if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
			}
		})
	}

	t.Run("GET command with deleted key", func(t *testing.T) {
		input := "GET foo"
		expOut := db.ErrKeyNotFound.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand("DEL foo", &buf, &ConnContext{})
		s.handleCommand(input, &buf, &ConnContext{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("COMPACT command with two key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut1 := "SET foo bar\n"
		expOut2 := "SET counter 13\n"

		var buf bytes.Buffer
		md := &mockDB{key: "multiple"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &ConnContext{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut1)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut1, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte(expOut2)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut2, buf.String())
		}
	})
}

func TestCommandParserWithMulti(t *testing.T) {
	t.Run("MULTI command with EXEC", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "GET foo", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", "QUEUED", "1) OK\n2) \"bar\""}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI command with DISCARD", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "GET foo", "DISCARD"}
		expOut := []string{MssgOK, "QUEUED", "QUEUED", MssgOK}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI command with EXEC with previous errors (invalid arguments)", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "INCRBY foo 5 6", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrWrongNumberOfArgs.Error(), ErrTranAbortedDueToPrevError.Error()}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI command with EXEC with previous errors (invalid command)", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "RANDOM NONSENSE", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrUnknownCommand.Error(), ErrTranAbortedDueToPrevError.Error()}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("EXEC command without MULTI", func(t *testing.T) {
		inputArr := []string{"EXEC"}
		expOut := []string{ErrExecWithoutMulti.Error()}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("DISCARD command without MULTI", func(t *testing.T) {
		inputArr := []string{"DISCARD"}
		expOut := []string{ErrDiscardWithoutMulti.Error()}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI calls nested", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo 5", "MULTI", "INCR foo", "GET foo", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrMultiCommandNested.Error(), "QUEUED", "QUEUED", "1) OK\n2) (integer) 6\n3) \"6\"\n"}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI & EXEC without any commands", func(t *testing.T) {
		inputArr := []string{"MULTI", "EXEC"}
		expOut := []string{MssgOK, MssgEmptyArray}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})

	t.Run("MULTI command with argument to EXEC", func(t *testing.T) {
		inputArr := []string{"MULTI", "SET foo bar", "EXEC GVK", "EXEC"}
		expOut := []string{MssgOK, "QUEUED", ErrWrongNumberOfArgs.Error(), ErrExecWithoutMulti.Error()}
		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		cc := &ConnContext{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, cc)

			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})
}

func TestServerConn(t *testing.T) {
	t.Run("Check if server is accepting connections ", func(t *testing.T) {
		// creates a test listener
		ln := bufconn.Listen(1024 * 1024)

		// start the test server with above listener
		md := &mockDB{}
		s := GetTestServer(md, ln)
		go s.Start()

		// trying to connect the above listener
		conn, err := ln.Dial()
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		conn.Close()
	})

	t.Run("Ping the server", func(t *testing.T) {
		buf := make([]byte, 1024)
		input := PING
		expOut := PONG

		// Assuming Server has a method Start that takes a net.Listener
		md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
		s := GetTestServer(md, ln)
		go s.Start()

		// starting connection
		conn, err := ln.Dial()
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		defer conn.Close()

		// Simulate client sending a command
		fmt.Fprintln(conn, input)

		// reading from connection
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Error while reading from connection: %v", err)
		}

		if !bytes.Contains(buf[:n], []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, string(buf[:n]))
		}
	})

	// TODO: writing test for disconnection of server pending
}

func TestConcurrentConn(t *testing.T) {
	t.Run("Check if server is accepting multiple connections ", func(t *testing.T) {
		// creates a test listener
		ln := bufconn.Listen(1024 * 1024)

		// start the test server
		md := &mockDB{}
		s := GetTestServer(md, ln)
		go s.Start()

		conn1, err := ln.Dial()
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}

		conn2, err := ln.Dial()
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}

		conn1.Close()
		conn2.Close()
	})

	t.Run("Check if multiple connections share the same storage", func(t *testing.T) {
		buf := make([]byte, 1024)
		testCases := []struct {
			input  []string
			expOut []string
		}{
			{[]string{"SET name John"}, []string{MssgOK}},
			{[]string{"GET name"}, []string{strconv.Quote("John")}},
		}

		// starting the server
		md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
		s := GetTestServer(md, ln)
		go s.Start()

		for _, tc := range testCases {
			// starting connection
			conn, err := ln.Dial()
			if err != nil {
				t.Fatalf("Failed to dial: %v", err)
			}
			defer conn.Close()

			// writing and verifying from connection
			for i, input := range tc.input {
				// writing to connection
				fmt.Fprintln(conn, input)

				// reading from connection
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatalf("Error while reading from connection: %v", err)
				}

				// verifying from connection
				if !bytes.Contains(buf[:n], []byte(tc.expOut[i])) {
					t.Errorf("Expected output to contain %q but got %s instead", tc.expOut[i], string(buf[:n]))
				}
			}
		}
	})

	t.Run("Check if multiple connection tran operations run independently", func(t *testing.T) {
		buf := make([]byte, 1024)
		testCases := []struct {
			input  []string
			expOut []string
		}{
			{[]string{"MULTI", "SET name John"}, []string{MssgOK, "QUEUED"}},
			{[]string{"INCR age"}, []string{"(integer) 1"}},
		}

		// starting the server
		md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
		s := GetTestServer(md, ln)
		go s.Start()

		for _, tc := range testCases {
			// starting connection
			conn, err := ln.Dial()
			if err != nil {
				t.Fatalf("Failed to dial: %v", err)
			}
			defer conn.Close()

			// writing and verifying from connection
			for i, input := range tc.input {
				// writing to connection
				fmt.Fprintln(conn, input)

				// reading from connection
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatalf("Error while reading from connection: %v", err)
				}

				// verifying from connection
				if !bytes.Contains(buf[:n], []byte(tc.expOut[i])) {
					t.Errorf("Expected output to contain %q but got %s instead", tc.expOut[i], string(buf[:n]))
				}
			}
		}

	})
}

func TestSelectCommand(t *testing.T) {

	t.Run("Check if multiple db indexes are running independently", func(t *testing.T) {
		buf := make([]byte, 1024)
		testCases := []struct {
			input  []string
			expOut []string
		}{
			{[]string{"SELECT 1", "SET name John"}, []string{MssgOK, MssgOK}},
			{[]string{"SELECT 2", "SET name Mills"}, []string{MssgOK, MssgOK}},
			{[]string{"SELECT 1", "GET name"}, []string{MssgOK, "John"}},
			{[]string{"SELECT 2", "GET name"}, []string{MssgOK, "Mills"}},
		}

		// starting the server
		md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
		s := GetTestServer(md, ln)
		// s := &Server{Listener: ln, Db: md}
		go s.Start()

		for _, tc := range testCases {
			// starting connection
			conn, err := ln.Dial()
			if err != nil {
				t.Fatalf("Failed to dial: %v", err)
			}
			defer conn.Close()

			// writing and verifying from connection
			for i, input := range tc.input {
				// writing to connection
				fmt.Fprintln(conn, input)

				// reading from connection
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatalf("Error while reading from connection: %v", err)
				}

				// verifying from connection
				if !bytes.Contains(buf[:n], []byte(tc.expOut[i])) {
					t.Errorf("Expected output to contain %q but got %s instead", tc.expOut[i], string(buf[:n]))
				}
			}
		}

	})
}

func TestStringSplit(t *testing.T) {
	md := &mockDB{}
	s := GetTestServer(md, nil)

	testCases := []struct {
		name    string
		input   string
		expOut  []string
		isError bool
		err     error
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
