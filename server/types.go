package server

import "fmt"

type Command struct {
	name string
	key  string
	val  string
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s %s", c.name, c.key, c.val)
}

type ConnContext struct {
	isMulti         bool      // to check if multi tran in progress
	multiCommands   []Command // to store commands of multi tran
	isTranDiscarded bool      // to check if multi tran was discarded
	dbIdx           int       // to store the db index
}
