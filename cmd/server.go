package cmd

import (
	"errors"
	"strconv"
	"fmt"
	"io"
	"regexp"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/redis"
)

type Command struct {
	name string
	key string
	val string
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s %s", c.name, c.key, c.val)
}

var ErrUnknownCommand = errors.New("unknown command")
var ErrWrongNumberOfArgs = errors.New("wrong number of arguments")
var ErrExecWithoutMulti = errors.New("exec without multi")
var ErrDiscardWithoutMulti = errors.New("discard without multi")
var ErrTranAbortedDueToPrevError = errors.New("transaction discarded because of previous errors")
var ErrMultiCommandNested = errors.New("multi calls can not be nested")
var MssgEmptyArray = "(empty array)"
var MssgOK = "OK"
// var ErrInvalidArguments = errors.New("invalid argument(s)")
// var ErrKeyNotFound = errors.New("(nil)")

const (
	GET string = "GET"
	SET string = "SET"
	DEL string = "DEL"
	INCR string = "INCR"
	INCRBY string = "INCRBY"
	MULTI string = "MULTI"
	QUEUED string = "QUEUED"
	EXEC string = "EXEC"
	DISCARD string = "DISCARD"
	COMPACT string = "COMPACT"
)


type Server struct {
	db redis.DbInterface
	out io.Writer
	isMulti bool
	multiCommandArr []Command
	isTranDiscarded bool
	// it will contain network related stuff later on
}

func(s *Server) handleCommand(input string) {
	// parse the input command
	i, err := StringSplit(input)
	if err != nil {
		fmt.Fprintln(s.out, err)
		return
	}

	// convert into command type
	c, err := s.makeCommand(i)
	if err != nil {
		fmt.Fprintln(s.out, err)
		return
	}

	// only add commands to multi tran if they aren't commands related to multi
	if s.isMulti && c.name != EXEC && c.name != DISCARD && c.name != MULTI {
		s.multiCommandArr = append(s.multiCommandArr, c)
		fmt.Fprintln(s.out, QUEUED)
		return
	}

	// take appropriate action
	s.takeAction(c)
}

func(s *Server) takeAction(c Command) {
	if c.name == SET {
		key := c.key
		val := c.val

		s.setAction(key, val)
	} else if c.name == GET {
		key := c.key

		s.getAction(key)
	} else if c.name == DEL {
		key := c.key

		s.delAction(key)
	} else if c.name == INCR {
		key := c.key

		s.incrAction(key)
	} else if c.name == INCRBY {
		key := c.key
		val := c.val

		s.incrbyAction(key, val)
	} else if c.name == MULTI {
		s.multiAction()
	} else if c.name == EXEC {
		s.execAction()
	} else if c.name == DISCARD {
		s.discardAction()
	} else if c.name == COMPACT {
		s.compactAction()
	} else {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", ErrUnknownCommand))
	}
}


func(s *Server) setAction(key, val string) {
	s.db.Set(key, val)
	fmt.Fprintln(s.out, MssgOK)
}

func(s *Server) getAction(key string) {
	val, err := s.db.Get(key)
	if err != nil {
		fmt.Fprint(s.out, redis.ErrKeyNotFound.Error())
		return
	}
	fmt.Fprintln(s.out, strconv.Quote(val))
}

func(s *Server) delAction(key string) {
	val := s.db.Del(key)
	fmt.Fprintln(s.out, val)
}

func(s *Server) incrAction(key string) {
	val, err := s.db.Incr(key)
	if err != nil {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", err))
		return
	}
	fmt.Fprintln(s.out, val)
}

func(s *Server) incrbyAction(key, val string) {
	val, err := s.db.Incrby(key, val)
	if err != nil {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", err))
		return
	}
	fmt.Fprintln(s.out, val)
}

func(s *Server) multiAction() {
	// if multi tran is already in progress
	if s.isMulti {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", ErrMultiCommandNested))
		return
	}
	
	s.isMulti = true
	fmt.Fprintln(s.out, MssgOK)
}

func(s *Server) execAction() {
	// can't exec without multi
	if !s.isMulti {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", ErrExecWithoutMulti))
		return
	}

	// if tran was discarded due to error
	if s.isTranDiscarded {
		s.isMulti = false
		s.isTranDiscarded = false
		s.multiCommandArr = []Command{}

		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", ErrTranAbortedDueToPrevError))
		return
	}

	// if no commands were given in a multi tran
	if len(s.multiCommandArr) == 0 {
		fmt.Fprintln(s.out, MssgEmptyArray)
		s.isMulti = false
		return
	}

	// normal execution
	s.isMulti = false
	for i, c := range s.multiCommandArr {
		fmt.Fprintf(s.out, "%d) ", i+1)
		s.takeAction(c)
	}
	s.multiCommandArr = []Command{}
}

func(s *Server) discardAction() {
	// can't discard without multi
	if !s.isMulti {
		fmt.Fprintln(s.out, fmt.Errorf("(error) ERR %v", ErrDiscardWithoutMulti))
		return
	}

	s.isMulti = false
	s.isTranDiscarded = false
	s.multiCommandArr = []Command{}
	fmt.Fprintln(s.out, MssgOK)
}

func(s *Server) compactAction() {
	data := s.db.GetAll()
	fmt.Println(len(data))
	if len(data) == 0 {
		fmt.Fprintf(s.out, "(nil)\n")
		return
	}

	for k, v := range data {
		fmt.Fprintf(s.out, "%s %s %s\n", SET, k, v)
	}	
}

func StringSplit(input string) ([]string, error) {
	isValid := isValidCommand(input)
	if !isValid {
		return nil, ErrUnknownCommand
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
			currentString = append(currentString, v)
		}
	}
	if len(currentString) > 0 {
        out = append(out, string(currentString))
    }
	return out, nil
}

func isValidCommand(command string) bool {
    var validCommandPattern = `^(?i)(?:"[A-Za-z0-9 ]+"|\b[A-Za-z0-9]+\b)(?:\s+"[^"]*"\s*|\s+\b[A-Za-z0-9]+\b\s*)*$`
    re := regexp.MustCompile(validCommandPattern)
    return re.MatchString(command)
}

func(s *Server) makeCommand(i []string) (Command, error) {
	if i[0] == "GET" || i[0] == "get" {
		if len(i) != 2 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: GET, key: i[1]}, nil
	} else if i[0] == "SET" || i[0] == "set" {
		if len(i) != 3 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: SET, key: i[1], val: i[2]}, nil
	} else if i[0] == "DEL" || i[0] == "del" {
		if len(i) != 2 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: DEL, key: i[1]}, nil
	} else if i[0] == "INCR" || i[0] == "incr" {
		if len(i) != 2 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: INCR, key: i[1]}, nil
	} else if i[0] == "INCRBY" || i[0] == "incrby" {
		if len(i) != 3 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: INCRBY, key: i[1], val: i[2]}, nil
	} else if i[0] == "MULTI" || i[0] == "multi" {
		if len(i) != 1 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: MULTI}, nil
	} else if i[0] == "EXEC" || i[0] == "exec" {
		if len(i) != 1 {
			s.isMulti = false
			s.isTranDiscarded = false
			s.multiCommandArr = []Command{}
			return Command{}, fmt.Errorf("(error) EXECABORT Transaction discarded because of: %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: EXEC}, nil
	} else if i[0] == "DISCARD" || i[0] == "discard" {
		if len(i) != 1 {
			if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: DISCARD}, nil
	} else if i[0] == "COMPACT" || i[0] == "compact" {
		if len(i) != 1 {
			// if s.isMulti {s.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: COMPACT}, nil
	}
	if s.isMulti {s.isTranDiscarded = true}
	return Command{}, fmt.Errorf("(error) ERR %v '%s', with args beginning with: ", ErrUnknownCommand, i[0])
}
