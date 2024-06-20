package server

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"bufio"
	"os"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
)

type Command struct {
	name string
	key string
	val string
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s %s", c.name, c.key, c.val)
}

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrWrongNumberOfArgs = errors.New("wrong number of arguments")
	ErrExecWithoutMulti = errors.New("exec without multi")
	ErrDiscardWithoutMulti = errors.New("discard without multi")
	ErrTranAbortedDueToPrevError = errors.New("transaction discarded because of previous errors")
	ErrMultiCommandNested = errors.New("multi calls can not be nested")
	MssgEmptyArray = "(empty array)"
	MssgOK = "OK"
	MssgNil = "(nil)"
)
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
	Db db.DbInterface
	Out io.Writer
	isMulti bool
	multiCommandArr []Command
	isTranDiscarded bool
	// it will contain network related stuff later on
}

func (s *Server) Start() {
    // Infinite loop to accept commands until an exit command is issued
    for {
        fmt.Fprint(s.Out, "> ")
        reader := bufio.NewReader(os.Stdin)
        input, err := reader.ReadString('\n')
        if err != nil {
            fmt.Fprintln(s.Out, "Error reading input:", err)
            continue
        }

        // Trim the newline character from the input
        input = strings.TrimSpace(input)

        // Check for exit command to break the loop
        if strings.ToLower(input) == "exit" {
            fmt.Fprintln(s.Out, "Exiting...")
            break
        }

        // Pass the command to the handleCommand method
        s.handleCommand(input)
    }
}

func(s *Server) handleCommand(input string) {
	// parse the input command
	i, err := StringSplit(input)
	if err != nil {
		fmt.Fprintln(s.Out, err)
		return
	}

	// convert into command type
	c, err := s.makeCommand(i)
	if err != nil {
		fmt.Fprintln(s.Out, err)
		return
	}

	// only add commands to multi tran if they aren't commands related to multi
	if s.isMulti && c.name != EXEC && c.name != DISCARD && c.name != MULTI {
		s.multiCommandArr = append(s.multiCommandArr, c)
		fmt.Fprintln(s.Out, QUEUED)
		return
	}

	// take appropriate action
	resp := s.takeAction(c)
	fmt.Fprintln(s.Out, resp)
}

func(s *Server) takeAction(c Command) string {
	switch c.name {
	case SET:
		return s.setAction(c.key, c.val)
	case GET:
		return s.getAction(c.key)
	case DEL:
		return s.delAction(c.key)
	case INCR:
		return s.incrAction(c.key)
	case INCRBY:
		return s.incrbyAction(c.key, c.val)
	case MULTI:
		return s.multiAction()
	case EXEC:
		return s.execAction()
	case DISCARD:
		return s.discardAction()
	case COMPACT:
		return s.compactAction()
	default:
		return fmt.Errorf("(error) ERR %v", ErrUnknownCommand).Error()
	}
}


func(s *Server) setAction(key, val string) string {
	s.Db.Set(key, val)
	return MssgOK
}

func(s *Server) getAction(key string) string {
	val, err := s.Db.Get(key)
	if err != nil {
		return db.ErrKeyNotFound.Error()
	}
	return strconv.Quote(val)
}

func(s *Server) delAction(key string) string {
	val := s.Db.Del(key)
	return val
}

func(s *Server) incrAction(key string) string {
	val, err := s.Db.Incr(key)
	if err != nil {
		return fmt.Errorf("(error) ERR %v", err).Error()
	}
	return val
}

func(s *Server) incrbyAction(key, val string) string {
	val, err := s.Db.Incrby(key, val)
	if err != nil {
		return fmt.Errorf("(error) ERR %v", err).Error()
	}
	return val
}

func(s *Server) multiAction() string {
	// if multi tran is already in progress
	if s.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrMultiCommandNested).Error()
	}
	
	s.isMulti = true
	return MssgOK
}

func(s *Server) execAction() string {
	// can't exec without multi
	if !s.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrExecWithoutMulti).Error()
	}

	// if tran was discarded due to error
	if s.isTranDiscarded {
		s.resetTran()
		return fmt.Errorf("(error) ERR %v", ErrTranAbortedDueToPrevError).Error()
	}

	// if no commands were given in a multi tran
	if len(s.multiCommandArr) == 0 {
		s.resetTran()
		return MssgEmptyArray
	}

	// normal execution
	var builder strings.Builder
	lastCmdArrIdx := len(s.multiCommandArr) - 1
	for i, c := range s.multiCommandArr {
		builder.WriteString(fmt.Sprintf("%d) ", i+1))

		// avoids the extra newline in final output for the last command in tran
		if lastCmdArrIdx == i {
			builder.WriteString(s.takeAction(c))
		} else {
			builder.WriteString(fmt.Sprintln(s.takeAction(c)))
		}
		
	}

	s.resetTran()
	return builder.String()
}

func(s *Server) discardAction() string {
	// can't discard without multi
	if !s.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrDiscardWithoutMulti).Error()
	}

	s.resetTran()
	return MssgOK
}

func(s *Server) compactAction() string {
	data := s.Db.GetAll()
	if len(data) == 0 {
		return MssgNil
	}

	var builder strings.Builder
	for k, v := range data {
		builder.WriteString(fmt.Sprintf("%s %s %s\n", SET, k, v))
	}
	return builder.String()
}

func(s *Server) resetTran() {
	s.isMulti = false
	s.isTranDiscarded = false
	s.multiCommandArr = []Command{}
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
			s.resetTran()
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
