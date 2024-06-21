package server

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"testing"
	"fmt"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
	"google.golang.org/grpc/test/bufconn"
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


func GetTestServer(db *mockDB, ln net.Listener) *Server {
	return &Server{
		Db: db,
		Listener: ln,
	}
}

func TestCommandParser(t *testing.T) {
	t.Run("SET command", func(t *testing.T) {
		input := "SET foo bar"
		expOut := MssgOK

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if md.key != "foo" {
			t.Errorf("Expected the key to be %q but didn't found it", md.key)
		}

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("SET command with invalid number of arguments (1)", func(t *testing.T) {
		input := "SET foo"
		expOut := ErrWrongNumberOfArgs.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("SET command with invalid number of arguments (3)", func(t *testing.T) {
		input := "SET foo bar extra"
		expOut := ErrWrongNumberOfArgs.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("GET command with valid key", func(t *testing.T) {
		input := "GET foo"
		expOut := strconv.Quote("bar")

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("GET command with deleted key", func(t *testing.T) {
		input := "GET foo"
		expOut := db.ErrKeyNotFound.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand("DEL foo", &buf, &TranState{})
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("GET command with invalid key", func(t *testing.T) {
		input := "GET foo"
		expOut := db.ErrKeyNotFound.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("GET command with invalid number of args (2)", func(t *testing.T) {
		input := "GET foo bar"
		expOut := ErrWrongNumberOfArgs.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("DEL command with valid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := db.DeleteSuccessMessage

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if md.key != "" {
			t.Errorf("Expected the key to be deleted but found %q", md.key)
		}

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("DEL command with invalid key", func(t *testing.T) {
		input := "DEL foo"
		expOut := db.DeleteFailedMessage

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCR command with valid key", func(t *testing.T) {
		input := "INCR foo"
		expOut := "(integer) 5"

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCR command with invalid key", func(t *testing.T) {
		input := "INCR foo"
		expOut := "(integer) 1"

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCR command with invalid value (string)", func(t *testing.T) {
		input := "INCR foo"
		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCR command with invalid value (float)", func(t *testing.T) {
		input := "INCR foo"
		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "10.5"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with valid key", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := "(integer) 9"

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid key", func(t *testing.T) {
		input := "INCRBY foo 8"
		expOut := "(integer) 8"

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid key and passed val is string", func(t *testing.T) {
		input := "INCRBY foo bar"
		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})
	t.Run("INCRBY command with passed key is string with integer val set", func(t *testing.T) {
		input := "INCRBY foo bar"

		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "4"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid value (string)", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "bar"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid value (float)", func(t *testing.T) {
		input := "INCRBY foo 5"
		expOut := db.ErrKeyNotInteger.Error()

		var buf bytes.Buffer
		md := &mockDB{key: "foo", val: "10.5"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid number of arguments (1)", func(t *testing.T) {
		input := "INCRBY foo"
		expOut := ErrWrongNumberOfArgs.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	t.Run("INCRBY command with invalid number of arguments (3)", func(t *testing.T) {
		input := "INCRBY foo bar extra"
		expOut := ErrWrongNumberOfArgs.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})

	
	t.Run("COMPACT command with one key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut := "SET foo bar\n"

		var buf bytes.Buffer
		md := &mockDB{key: "one"}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

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
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut1)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut1, buf.String())
		}
		if !bytes.Contains(buf.Bytes(), []byte(expOut2)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut2, buf.String())
		}
	})

	t.Run("COMPACT command with no key-val pair", func(t *testing.T) {
		input := "COMPACT"
		expOut := "(nil)\n"

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
		}
	})


	t.Run("Invalid command", func(t *testing.T) {
		input := "gibberish foo bar"
		expOut := ErrUnknownCommand.Error()

		var buf bytes.Buffer
		md := &mockDB{}
		s := GetTestServer(md, nil)
		s.handleCommand(input, &buf, &TranState{})

		if !bytes.Contains(buf.Bytes(), []byte(expOut)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut, buf.String())
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
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
		ts := &TranState{}

		for i, input := range inputArr {
			s.handleCommand(input, &buf, ts)

			
			if !bytes.Contains(buf.Bytes(), []byte(expOut[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut[i], buf.String())
			}

			buf.Reset()
		}
	})
}

func TestServerConn(t *testing.T){
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
        s := &Server{Listener: ln, Db: md}
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

	// writing test for disconnection of server pending
	// t.Run("disconnect the server", func(t *testing.T) {
    //     buf := make([]byte, 1024)
    //     input := DISCONNECT
    //     // expOut := PONG


    //     // Assuming Server has a method Start that takes a net.Listener
    //     md := &mockDB{}
	// 	ln := bufconn.Listen(1024 * 1024)
    //     s := &Server{Listener: ln, Db: md}
    //     go s.Start()
    
    //     // starting connection
    //     conn, err := ln.Dial()
    //     if err != nil {
    //         t.Fatalf("Failed to dial: %v", err)
    //     }
    //     defer conn.Close()

    //     // Simulate client sending a command
    //     fmt.Fprintln(conn, input)

    //     // trying to connect on same connection after connection
	// 	fmt.Fprintln(conn, PING)
	// 	n, err := conn.Read(buf)
	// 	// if err == nil {
	// 	// 	t.Fatalf("Expected error but got nil")
	// 	// }

	// 	fmt.Println(string(buf[:n]))
        

    //     // if !bytes.Contains(buf[:n], []byte(expOut)) {
	// 	// 	t.Errorf("Expected output to contain %q but got %s instead", expOut, string(buf[:n]))
	// 	// }
    // })
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
        input1 := "SET name John"
		expOut1 := MssgOK + "\n"
        input2 := "GET name"
		expOut2 := strconv.Quote("John") + "\n"


        // starting the server
        md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
        s := &Server{Listener: ln, Db: md}
        go s.Start()
    
        // starting 1st connection
        conn1, err := ln.Dial()
        if err != nil {
            t.Fatalf("Failed to dial: %v", err)
        }
        defer conn1.Close()

        // sending 1st command
        fmt.Fprintln(conn1, input1)

        // reading & verifying from connection1
        n, err := conn1.Read(buf)
        if err != nil {
            t.Fatalf("Error while reading from connection: %v", err)
        }

        if !bytes.Contains(buf[:n], []byte(expOut1)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut1, string(buf[:n]))
		}

        // starting 2nd connection
        conn2, err := ln.Dial()
        if err != nil {
            t.Fatalf("Failed to dial: %v", err)
        }
        defer conn2.Close()

        // sending 1st command
        fmt.Fprintln(conn2, input2)

        // reading & verifying from connection2
        n, err = conn2.Read(buf)
        if err != nil {
            t.Fatalf("Error while reading from connection: %v", err)
        }

        if !bytes.Contains(buf[:n], []byte(expOut2)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut2, string(buf[:n]))
		}
    })

	t.Run("Check if multiple connections multi tran operations run independently", func(t *testing.T) {
        buf := make([]byte, 1024)
        input1 := []string{"MULTI", "SET name John"}
		expOut1 := []string{MssgOK, QUEUED}
        input2 := "INCR age"
		expOut2 := "(integer) 1"
		


        // starting the server
        md := &mockDB{}
		ln := bufconn.Listen(1024 * 1024)
        s := &Server{Listener: ln, Db: md}
        go s.Start()
    
        // starting 1st connection
        conn1, err := ln.Dial()
        if err != nil {
            t.Fatalf("Failed to dial: %v", err)
        }
        defer conn1.Close()

		// writing and verifying from 1st connection
		for i, input := range input1 {
			fmt.Fprintln(conn1, input)
			
			n, err := conn1.Read(buf)
			if err != nil {
				t.Fatalf("Error while reading from connection: %v", err)
			}
	
			if !bytes.Contains(buf[:n], []byte(expOut1[i])) {
				t.Errorf("Expected output to contain %q but got %s instead", expOut1[i], string(buf[:n]))
			}
		}
    
        // starting 2nd connection
        conn2, err := ln.Dial()
        if err != nil {
            t.Fatalf("Failed to dial: %v", err)
        }
        defer conn2.Close()

        // sending 1st command
        fmt.Fprintln(conn2, input2)

        // reading & verifying from connection2
        n, err := conn2.Read(buf)
        if err != nil {
            t.Fatalf("Error while reading from connection: %v", err)
        }

        if !bytes.Contains(buf[:n], []byte(expOut2)) {
			t.Errorf("Expected output to contain %q but got %s instead", expOut2, string(buf[:n]))
		}
    })
}