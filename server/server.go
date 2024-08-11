package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

var (
	ErrUnknownCommand            = errors.New("unknown command")
	ErrWrongNumberOfArgs         = errors.New("wrong number of arguments")
	ErrExecWithoutMulti          = errors.New("exec without multi")
	ErrDiscardWithoutMulti       = errors.New("discard without multi")
	ErrTranAbortedDueToPrevError = errors.New("transaction discarded because of previous errors")
	ErrMultiCommandNested        = errors.New("multi calls can not be nested")
	ErrDBIndexOutOfRange         = errors.New("(error) ERR DB index is out of range")
	ErrKeyNotFound = errors.New("failed to find the key")
	MssgEmptyArray               = "(empty array)"
	MssgOK                       = "OK"
	MssgNil                      = "(nil)"
	DbRangeMin                   = 0
	DbRangeMax                   = 15
)

const (
	GET        string = "GET"
	SET        string = "SET"
	DEL        string = "DEL"
	INCR       string = "INCR"
	INCRBY     string = "INCRBY"
	MULTI      string = "MULTI"
	QUEUED     string = "QUEUED"
	EXEC       string = "EXEC"
	DISCARD    string = "DISCARD"
	COMPACT    string = "COMPACT"
	PING       string = "PING"
	PONG       string = "PONG"
	DISCONNECT string = "DISCONNECT"
	SELECT     string = "SELECT"
)

type Command struct {
	name string
	key  string
	val  string
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s %s", c.name, c.key, c.val)
}

type ConnContext struct {
	isMulti         bool	// to check if multi tran in progress
	multiCommandArr []Command	// to store commands of multi tran
	isTranDiscarded bool	// to check if multi tran was discarded
	dbIdx           int		// to store the db index
}

type Server struct {
	Db       map[int]db.DbInterface
	Listener net.Listener
}

// starts the server
// entrypoint for the app
func (s *Server) Start() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Fprintf(conn, "Error while accepting connection: %v", err)
		}

		// launching new go routine for each connection
		go s.handleConnection(conn, &ConnContext{})
	}
}

// start method for cli application
// func (s *Server) Start() {
//     // infinite loop to accept commands until an exit command is issued
//     for {
//         fmt.Fprint(s.Out, "> ")
//         reader := bufio.NewReader(os.Stdin)
//         input, err := reader.ReadString('\n')
//         if err != nil {
//             fmt.Fprintln(s.Out, "Error reading input:", err)
//             continue
//         }

//         // trim the newline character from the input
//         input = strings.TrimSpace(input)

//         // check for exit command to break the loop
//         if strings.ToLower(input) == "exit" {
//             fmt.Fprintln(s.Out, "Exiting...")
//             break
//         }

//         // pass the command to the handleCommand method
//         s.handleCommand(input)
//     }
// }

// reads from the connection until conn is terminated
func (s *Server) handleConnection(conn net.Conn, cc *ConnContext) {
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

		// starts acting on the command
		s.handleCommand(string(buf[:n]), conn, cc)
	}
}

// parses the command and takes action
func (s *Server) handleCommand(input string, out io.Writer, cc *ConnContext) {
	// parse the input command
	i, err := s.stringSplit(input)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	// convert raw command into command type
	c, err := s.makeCommand(i, cc)
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

	// only add commands to multi tran isMulti is ON & if they aren't commands related to multi
	if cc.isMulti && c.name != EXEC && c.name != DISCARD && c.name != MULTI {
		cc.multiCommandArr = append(cc.multiCommandArr, c)
		fmt.Fprintln(out, QUEUED)
		return
	}

	// take appropriate action
	resp := s.takeAction(c, cc)
	fmt.Fprintln(out, resp)
}

// takes action based on the command name
func (s *Server) takeAction(c Command, cc *ConnContext) string {
	switch c.name {
	case PING:
		return s.pingAction()
	case SELECT:
		return s.selectAction(cc, c.val)
	case SET:
		return s.setAction(cc, c.key, c.val)
	case GET:
		return s.getAction(cc, c.key)
	case DEL:
		return s.delAction(cc, c.key)
	case INCR:
		return s.incrAction(cc, c.key)
	case INCRBY:
		return s.incrbyAction(cc, c.key, c.val)
	case MULTI:
		return s.multiAction(cc)
	case EXEC:
		return s.execAction(cc)
	case DISCARD:
		return s.discardAction(cc)
	case COMPACT:
		return s.compactAction(cc)
	default:
		return fmt.Errorf("(error) ERR %v", ErrUnknownCommand).Error()
	}
}

func (s *Server) pingAction() string {
	return PONG
}

func (s *Server) selectAction(cc *ConnContext, val string) string {
	i, err := strconv.Atoi(val)
	if err != nil {
		return db.ErrKeyNotInteger.Error()
	}

	// db index should be between 0 and 15
	if i < DbRangeMin || i > DbRangeMax {
		return ErrDBIndexOutOfRange.Error()
	}

	// checking for the particular db index
	// create db if its not there and set the index
	_, ok := s.Db[i]
	if !ok {
		s.Db[i] = db.GetNewDB(inMemoryStore.NewInMemoryStore())
	}
	cc.dbIdx = i

	return MssgOK
}

func (s *Server) setAction(cc *ConnContext, key, val string) string {
	s.Db[cc.dbIdx].Set(key, val)
	return MssgOK
}

func (s *Server) getAction(cc *ConnContext, key string) string {
	val, err := s.Db[cc.dbIdx].Get(key)
	if err != nil {
		return db.ErrKeyNotFound.Error()
	}
	return strconv.Quote(val)
}

func (s *Server) delAction(cc *ConnContext, key string) string {
	val := s.Db[cc.dbIdx].Del(key)
	return val
}

func (s *Server) incrAction(cc *ConnContext, key string) string {
	val, err := s.Db[cc.dbIdx].Incr(key)
	if err != nil {
		return fmt.Errorf("(error) ERR %v", err).Error()
	}
	return val
}

func (s *Server) incrbyAction(cc *ConnContext, key, val string) string {
	val, err := s.Db[cc.dbIdx].Incrby(key, val)
	if err != nil {
		return fmt.Errorf("(error) ERR %v", err).Error()
	}
	return val
}

func (s *Server) multiAction(cc *ConnContext) string {
	// if multi tran is already in progress
	if cc.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrMultiCommandNested).Error()
	}

	cc.isMulti = true
	return MssgOK
}

func (s *Server) execAction(cc *ConnContext) string {
	// can't exec without multi
	if !cc.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrExecWithoutMulti).Error()
	}

	// if tran was discarded due to error
	if cc.isTranDiscarded {
		s.resetTran(cc)
		return fmt.Errorf("(error) ERR %v", ErrTranAbortedDueToPrevError).Error()
	}

	// if no commands were given in a multi tran
	if len(cc.multiCommandArr) == 0 {
		s.resetTran(cc)
		return MssgEmptyArray
	}

	// normal execution
	var builder strings.Builder
	lastCmdArrIdx := len(cc.multiCommandArr) - 1
	for i, c := range cc.multiCommandArr {
		builder.WriteString(fmt.Sprintf("%d) ", i+1))

		// avoids the extra newline in final output for the last command in tran
		if lastCmdArrIdx == i {
			builder.WriteString(s.takeAction(c, cc))
		} else {
			builder.WriteString(fmt.Sprintln(s.takeAction(c, cc)))
		}

	}

	s.resetTran(cc)
	return builder.String()
}

func (s *Server) discardAction(cc *ConnContext) string {
	// can't discard without multi
	if !cc.isMulti {
		return fmt.Errorf("(error) ERR %v", ErrDiscardWithoutMulti).Error()
	}

	s.resetTran(cc)
	return MssgOK
}

func (s *Server) resetTran(cc *ConnContext) {
	cc.isMulti = false
	cc.isTranDiscarded = false
	cc.multiCommandArr = []Command{}
}

func (s *Server) compactAction(cc *ConnContext) string {
	data := s.Db[cc.dbIdx].GetAll()
	if len(data) == 0 {
		return MssgNil
	}

	var builder strings.Builder
	for k, v := range data {
		builder.WriteString(fmt.Sprintf("%s %s %s\n", SET, k, v))
	}
	return builder.String()
}

func (s *Server) stringSplit(input string) ([]string, error) {
	trimmedInput := strings.TrimSpace(input)
	isValid := s.isValidCommand(trimmedInput)
	if !isValid {
		return nil, ErrUnknownCommand
	}

	// declaration of single quote, double quote and space character vars
	var (
		sq rune = '\''
		dq rune = '"'
		sp rune = ' '
	)

	out := []string{}
	currentString := []rune{}
	isInsideQuote := false

	for _, v := range trimmedInput {
		// if char is quote, flip the isInsideQuote flag
		if v == sq || v == dq {
			isInsideQuote = !isInsideQuote
		}

		// if char is space and not inside quote, append the string to out
		if v == sp && !isInsideQuote {
			out = append(out, string(currentString))
			currentString = []rune{}
			continue
		}

		// if char isn't quote, append the char to currentString
		if v != sq && v != dq {
			currentString = append(currentString, v)
		}
	}
	// append the last string to out
	if len(currentString) > 0 {
		out = append(out, string(currentString))
	}

	return out, nil
}

func (s *Server) isValidCommand(command string) bool {
	// regex pattern for valid command
	var validCommandPattern = `^(?i)(?:"[A-Za-z0-9 ]+"|\b[A-Za-z0-9]+\b)(?:\s+"[^"]*"\s*|\s+\b[A-Za-z0-9]+\b\s*)*$`
	re := regexp.MustCompile(validCommandPattern)
	return re.MatchString(command)
}

// turns the raw command into Command type
// returns error if command is unknown or invalid number of args
func (s *Server) makeCommand(i []string, cc *ConnContext) (Command, error) {
	switch {
	case i[0] == "SELECT" || i[0] == "select":
		if len(i) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: SELECT, val: i[1]}, nil
	case i[0] == "PING" || i[0] == "ping":
		if len(i) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: PING}, nil
	case i[0] == "GET" || i[0] == "get":
		if len(i) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: GET, key: i[1]}, nil
	case i[0] == "SET" || i[0] == "set":
		if len(i) != 3 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: SET, key: i[1], val: i[2]}, nil
	case i[0] == "DEL" || i[0] == "del":
		if len(i) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: DEL, key: i[1]}, nil
	case i[0] == "INCR" || i[0] == "incr":
		if len(i) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: INCR, key: i[1]}, nil
	case i[0] == "INCRBY" || i[0] == "incrby":
		if len(i) != 3 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: INCRBY, key: i[1], val: i[2]}, nil
	case i[0] == "MULTI" || i[0] == "multi":
		if len(i) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: MULTI}, nil
	case i[0] == "EXEC" || i[0] == "exec":
		if len(i) != 1 {
			s.resetTran(cc)
			return Command{}, fmt.Errorf("(error) EXECABORT Transaction discarded because of: %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: EXEC}, nil
	case i[0] == "DISCARD" || i[0] == "discard":
		if len(i) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: DISCARD}, nil
	case i[0] == "COMPACT" || i[0] == "compact":
		if len(i) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: COMPACT}, nil
	case i[0] == "DISCONNECT" || i[0] == "disconnect":
		if len(i) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, i[0])
		}
		return Command{name: DISCONNECT}, nil
	default:
		if cc.isMulti {
			cc.isTranDiscarded = true
		}
		return Command{}, fmt.Errorf("(error) ERR %v '%s', with args beginning with: ", ErrUnknownCommand, i[0])
	}	
}
