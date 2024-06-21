package server

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"net"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
)

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
	PING string = "PING"
	PONG string = "PONG"
	DISCONNECT string = "DISCONNECT"
)

type Command struct {
	name string
	key string
	val string
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s %s", c.name, c.key, c.val)
}

type TranState struct {
	isMulti         bool
	multiCommandArr []Command
	isTranDiscarded bool
}

type Server struct {
	Db db.DbInterface
	// isMulti bool
	// multiCommandArr []Command
	// isTranDiscarded bool
	// it will contain network related stuff later on
	Listener net.Listener
}

func(s *Server) Start() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Fprintf(conn, "Error while accepting connection: %v", err)
		}
		
		// launching new go routine for each connection
		go s.handleConnection(conn, &TranState{})
	}
}

// for cli application
// func (s *Server) Start() {
//     // Infinite loop to accept commands until an exit command is issued
//     for {
//         fmt.Fprint(s.Out, "> ")
//         reader := bufio.NewReader(os.Stdin)
//         input, err := reader.ReadString('\n')
//         if err != nil {
//             fmt.Fprintln(s.Out, "Error reading input:", err)
//             continue
//         }

//         // Trim the newline character from the input
//         input = strings.TrimSpace(input)

//         // Check for exit command to break the loop
//         if strings.ToLower(input) == "exit" {
//             fmt.Fprintln(s.Out, "Exiting...")
//             break
//         }

//         // Pass the command to the handleCommand method
//         s.handleCommand(input)
//     }
// }


func(s *Server) handleConnection(conn net.Conn, ts *TranState) {
	defer conn.Close()

	buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            if err != io.EOF {
                fmt.Println("Error reading:", err.Error())
            }
            break
        }

		s.handleCommand(string(buf[:n]), conn, ts)
    }
}

func(s *Server) handleCommand(input string, out io.Writer, ts *TranState) {
	// parse the input command
	i, err := s.stringSplit(input)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	// convert into command type
	c, err := s.makeCommand(i, ts)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	// handling disconnect
	if c.name == DISCONNECT {
		if conn, ok := out.(net.Conn); ok {
		    conn.Close()
		    return
		}
	}

	// only add commands to multi tran if they aren't commands related to multi
	if ts.isMulti && c.name != EXEC && c.name != DISCARD && c.name != MULTI {
		ts.multiCommandArr = append(ts.multiCommandArr, c)
		fmt.Fprintln(out, QUEUED)
		return
	}

	// take appropriate action
	resp := s.takeAction(c, ts)
	fmt.Fprintln(out, resp)
}

func(s *Server) takeAction(c Command, ts *TranState) string {
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
		return s.multiAction(ts)
	case EXEC:
		return s.execAction(ts)
	case DISCARD:
		return s.discardAction(ts)
	case COMPACT:
		return s.compactAction()
	case PING:
		return s.pingAction()
	default:
		return fmt.Errorf("(error) ERR %v", ErrUnknownCommand).Error()
	}
}

func (s *Server) pingAction() string {
	return PONG
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

func(s *Server) multiAction(ts *TranState) string {
	// if multi tran is already in progress
	if ts.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrMultiCommandNested).Error()
	}
	
	ts.isMulti = true
	return MssgOK
}

func(s *Server) execAction(ts *TranState) string {
	// can't exec without multi
	if !ts.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrExecWithoutMulti).Error()
	}

	// if tran was discarded due to error
	if ts.isTranDiscarded {
		s.resetTran(ts)
		return fmt.Errorf("(error) ERR %v", ErrTranAbortedDueToPrevError).Error()
	}

	// if no commands were given in a multi tran
	if len(ts.multiCommandArr) == 0 {
		s.resetTran(ts)
		return MssgEmptyArray
	}

	// normal execution
	var builder strings.Builder
	lastCmdArrIdx := len(ts.multiCommandArr) - 1
	for i, c := range ts.multiCommandArr {
		builder.WriteString(fmt.Sprintf("%d) ", i+1))

		// avoids the extra newline in final output for the last command in tran
		if lastCmdArrIdx == i {
			builder.WriteString(s.takeAction(c, ts))
		} else {
			builder.WriteString(fmt.Sprintln(s.takeAction(c, ts)))
		}
		
	}

	s.resetTran(ts)
	return builder.String()
}

func(s *Server) discardAction(ts *TranState) string {
	// can't discard without multi
	if !ts.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrDiscardWithoutMulti).Error()
	}

	s.resetTran(ts)
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

func(s *Server) resetTran(ts *TranState) {
	ts.isMulti = false
	ts.isTranDiscarded = false
	ts.multiCommandArr = []Command{}
}

func(s *Server) stringSplit(input string) ([]string, error) {
	trimmedInput := strings.TrimSpace(input)
	isValid := s.isValidCommand(trimmedInput)
	if !isValid {
		return nil, ErrUnknownCommand
	}

	var sq rune = '\''
	var dq rune = '"'
	var sp rune = ' '

	out := []string{}
	currentString := []rune{}
	isInsideQuote := false

	for _, v := range trimmedInput {
		if v == sq || v == dq {
			isInsideQuote = !isInsideQuote
		}
		if v == sp && !isInsideQuote {
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

func(s *Server) isValidCommand(command string) bool {
    var validCommandPattern = `^(?i)(?:"[A-Za-z0-9 ]+"|\b[A-Za-z0-9]+\b)(?:\s+"[^"]*"\s*|\s+\b[A-Za-z0-9]+\b\s*)*$`
    re := regexp.MustCompile(validCommandPattern)
    return re.MatchString(command)
}

func(s *Server) makeCommand(i []string, ts *TranState) (Command, error) {
	if i[0] == "GET" || i[0] == "get" {
		if len(i) != 2 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: GET, key: i[1]}, nil
	} else if i[0] == "SET" || i[0] == "set" {
		if len(i) != 3 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: SET, key: i[1], val: i[2]}, nil
	} else if i[0] == "DEL" || i[0] == "del" {
		if len(i) != 2 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: DEL, key: i[1]}, nil
	} else if i[0] == "INCR" || i[0] == "incr" {
		if len(i) != 2 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: INCR, key: i[1]}, nil
	} else if i[0] == "INCRBY" || i[0] == "incrby" {
		if len(i) != 3 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: INCRBY, key: i[1], val: i[2]}, nil
	} else if i[0] == "MULTI" || i[0] == "multi" {
		if len(i) != 1 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: MULTI}, nil
	} else if i[0] == "EXEC" || i[0] == "exec" {
		if len(i) != 1 {
			s.resetTran(ts)
			return Command{}, fmt.Errorf("(error) EXECABORT Transaction discarded because of: %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: EXEC}, nil
	} else if i[0] == "DISCARD" || i[0] == "discard" {
		if len(i) != 1 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: DISCARD}, nil
	} else if i[0] == "COMPACT" || i[0] == "compact" {
		if len(i) != 1 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}	
		return Command{name: COMPACT}, nil
	} else if i[0] == "PING" || i[0] == "ping" {
		if len(i) != 1 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: PING}, nil
	} else if i[0] == "DISCONNECT" || i[0] == "disconnect" {
		if len(i) != 1 {
			if ts.isMulti {ts.isTranDiscarded = true}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: DISCONNECT}, nil
	}
	if ts.isMulti {ts.isTranDiscarded = true}
	return Command{}, fmt.Errorf("(error) ERR %v '%s', with args beginning with: ", ErrUnknownCommand, i[0])
}
