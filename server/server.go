package server

import (
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/justsushant/one2n-go-bootcamp/go-redis/db"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/db/keyvaldb"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/store/inmemorystore"
)

type Server struct {
	database map[int]db.Database
	listener net.Listener
}

func NewServer(dbx db.Database, ln net.Listener) *Server {
	return &Server{
		database: map[int]db.Database{0: dbx},
		listener: ln,
	}
}

// starts the server
// entrypoint for the app
func (s *Server) Start() {
	for {
		conn, err := s.listener.Accept()
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
				fmt.Fprintln(conn, "Error reading:", err.Error())
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
	inputParts, err := s.parseInput(input)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	// convert raw command into command type
	cmd, err := s.makeCommand(inputParts, cc)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	// handling disconnect
	if cmd.name == DISCONNECT {
		if conn, ok := out.(net.Conn); ok {
			conn.Close()
			return
		}
	}

	// only add commands to multi tran if isMulti is ON & if they aren't commands related to multi
	if cc.isMulti && cmd.name != EXEC && cmd.name != DISCARD && cmd.name != MULTI {
		cc.multiCommands = append(cc.multiCommands, cmd)
		fmt.Fprintln(out, QUEUED)
		return
	}

	// take appropriate action
	resp := s.takeAction(cc, cmd)
	fmt.Fprintln(out, resp)
}

// takes action based on the command name
func (s *Server) takeAction(cc *ConnContext, cmd Command) string {
	switch cmd.name {
	case PING:
		return s.pingAction()
	case SELECT:
		return s.selectAction(cc, cmd.val)
	case SET:
		return s.setAction(cc, cmd.key, cmd.val)
	case GET:
		return s.getAction(cc, cmd.key)
	case DEL:
		return s.delAction(cc, cmd.key)
	case INCR:
		return s.incrAction(cc, cmd.key)
	case INCRBY:
		return s.incrbyAction(cc, cmd.key, cmd.val)
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

func (s *Server) selectAction(cc *ConnContext, dbIdx string) string {
	dbIdxInt, err := strconv.Atoi(dbIdx)
	if err != nil {
		return db.ErrKeyNotInteger.Error()
	}

	// db index should be between 0 and 15
	if dbIdxInt < DB_RANGE_MIN || dbIdxInt > DB_RANGE_MAX {
		return ErrDBIndexOutOfRange.Error()
	}

	// checking for the particular db index
	// create db if its not there and set the index
	_, ok := s.database[dbIdxInt]
	if !ok {
		s.database[dbIdxInt] = keyvaldb.NewDB(inmemorystore.NewStore())
	}
	cc.dbIdx = dbIdxInt

	return MSSG_OK
}

func (s *Server) setAction(cc *ConnContext, key, val string) string {
	s.database[cc.dbIdx].Set(key, val)
	return MSSG_OK
}

func (s *Server) getAction(cc *ConnContext, key string) string {
	val, err := s.database[cc.dbIdx].Get(key)
	if err != nil {
		return db.ErrKeyNotFound.Error()
	}
	return strconv.Quote(val)
}

func (s *Server) delAction(cc *ConnContext, key string) string {
	val := s.database[cc.dbIdx].Del(key)
	return val
}

func (s *Server) incrAction(cc *ConnContext, key string) string {
	val, err := s.database[cc.dbIdx].Incr(key)
	if err != nil {
		return fmt.Errorf("(error) ERR %v", err).Error()
	}
	return val
}

func (s *Server) incrbyAction(cc *ConnContext, key, val string) string {
	val, err := s.database[cc.dbIdx].Incrby(key, val)
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
	return MSSG_OK
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
	if len(cc.multiCommands) == 0 {
		s.resetTran(cc)
		return MSSG_EMPTY_ARRAY
	}

	// normal execution
	var builder strings.Builder
	lastCmdArrIdx := len(cc.multiCommands) - 1
	for i, c := range cc.multiCommands {
		builder.WriteString(fmt.Sprintf("%d) ", i+1))

		// avoids the extra newline in final output for the last command in tran
		if lastCmdArrIdx == i {
			builder.WriteString(s.takeAction(cc, c))
		} else {
			builder.WriteString(fmt.Sprintln(s.takeAction(cc, c)))
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
	return MSSG_OK
}

func (s *Server) resetTran(cc *ConnContext) {
	cc.isMulti = false
	cc.isTranDiscarded = false
	cc.multiCommands = []Command{}
}

func (s *Server) compactAction(cc *ConnContext) string {
	data := s.database[cc.dbIdx].GetAll()
	if len(data) == 0 {
		return MSSG_NIL
	}

	// string builder implementation
	// var builder strings.Builder
	// for k, v := range data {
	// 	builder.WriteString(fmt.Sprintf("%s %s %s\n", SET, k, v))
	// }
	// return builder.String()

	var dataArr []string
	for k, v := range data {
		dataArr = append(dataArr, fmt.Sprintf("%s %s %s", SET, k, v))
	}
	return strings.Join(dataArr, "\n")
}

func (s *Server) parseInput(input string) ([]string, error) {
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

func (s *Server) isValidCommand(cmd string) bool {
	// regex pattern for valid command
	var validCmdPattern = `^(?i)(?:"[A-Za-z0-9 ]+"|\b[A-Za-z0-9]+\b)(?:\s+"[^"]*"\s*|\s+\b[A-Za-z0-9]+\b\s*)*$`
	re := regexp.MustCompile(validCmdPattern)
	return re.MatchString(cmd)
}

// turns the raw command into Command type
// returns error if command is unknown or invalid number of args
func (s *Server) makeCommand(cmdParts []string, cc *ConnContext) (Command, error) {
	cmdName := strings.ToLower(cmdParts[0])

	switch {
	case cmdName == "select":
		if len(cmdParts) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: SELECT, val: cmdParts[1]}, nil
	case cmdName == "ping":
		if len(cmdParts) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: PING}, nil
	case cmdName == "get":
		if len(cmdParts) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: GET, key: cmdParts[1]}, nil
	case cmdName == "set":
		if len(cmdParts) != 3 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: SET, key: cmdParts[1], val: cmdParts[2]}, nil
	case cmdName == "del":
		if len(cmdParts) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: DEL, key: cmdParts[1]}, nil
	case cmdName == "incr":
		if len(cmdParts) != 2 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: INCR, key: cmdParts[1]}, nil
	case cmdName == "incrby":
		if len(cmdParts) != 3 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: INCRBY, key: cmdParts[1], val: cmdParts[2]}, nil
	case cmdName == "multi":
		if len(cmdParts) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: MULTI}, nil
	case cmdName == "exec":
		if len(cmdParts) != 1 {
			s.resetTran(cc)
			return Command{}, fmt.Errorf("(error) EXECABORT Transaction discarded because of: %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: EXEC}, nil
	case cmdName == "discard":
		if len(cmdParts) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: DISCARD}, nil
	case cmdName == "compact":
		if len(cmdParts) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: COMPACT}, nil
	case cmdName == "disconnect":
		if len(cmdParts) != 1 {
			if cc.isMulti {
				cc.isTranDiscarded = true
			}
			return Command{}, fmt.Errorf("(error) ERR %v for '%s' command", ErrWrongNumberOfArgs, cmdName)
		}
		return Command{name: DISCONNECT}, nil
	default:
		if cc.isMulti {
			cc.isTranDiscarded = true
		}
		return Command{}, fmt.Errorf("(error) ERR %v '%s', with args beginning with: ", ErrUnknownCommand, cmdName)
	}
}
