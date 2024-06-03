package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/redis"
)

type Command struct {
	name string
	key string
	val string
}

var ErrInvalidCommand = errors.New("invalid command")

const (
	GET string = "GET"
	SET string = "SET"
	DEL string = "DEL"
)


type Server struct {
	db redis.DbInterface
	out io.Writer
	// it will contain network related stuff later on
}

func(s *Server) ParseCommand(input string) {
	c := strings.Split(input, " ")

	if c[0] == SET {
		key := c[1]
		val := c[2]

		s.SetAction(key, val)
	} else if c[0] == GET {
		key := c[1]

		s.GetAction(key)
	} else if c[0] == DEL {
		key := c[1]

		s.DelAction(key)
	} else {
		fmt.Fprintln(s.out, "Invalid Command")
	}
}

func(s *Server) SetAction(key, val string) {
	s.db.Set(key, val)
	fmt.Fprintln(s.out, "OK")
}

func(s *Server) GetAction(key string) {
	val, err := s.db.Get(key)
	if err != nil {
		if errors.Is(err, redis.ErrKeyNotFound) {
			fmt.Fprint(s.out, redis.ErrKeyNotFound.Error())
			return
		}
		fmt.Fprintf(s.out, "Unexpected Error: %v\n", err)
	}

	fmt.Fprintln(s.out, val)
}

func(s *Server) DelAction(key string) {
	val := s.db.Del(key)
	fmt.Fprintln(s.out, val)
}

func StringSplit(input string) ([]string, error) {
	if !strings.ContainsRune(input, '"') {
		return strings.Split(input, " "), nil
	} 
	var sq rune = '\''
	var dq rune = '"'
	var s rune = ' '
	out := []string{}
	currentString := []rune{}
	isInsideQuote := false
	for _, v := range input {
		if v == sq || v == dq {
			isInsideQuote = !isInsideQuote
		}
		if v == s && !isInsideQuote {
			out = append(out, string(currentString))
			currentString = []rune{}
			continue
		}
		if v != sq && v != dq {
			// fmt.Println(out, string(currentString), string(v))
			currentString = append(currentString, v)
		}
		// if !isInsideQuote && (v == sq || v == dq) {
		// 	out = append(out, string(currentString))
		// }
	}
	if len(currentString) > 0 {
        out = append(out, string(currentString))
    }
	return out, nil
}