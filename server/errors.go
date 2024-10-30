package server

import "errors"

var (
	ErrUnknownCommand            = errors.New("unknown command")
	ErrWrongNumberOfArgs         = errors.New("wrong number of arguments")
	ErrExecWithoutMulti          = errors.New("exec without multi")
	ErrDiscardWithoutMulti       = errors.New("discard without multi")
	ErrTranAbortedDueToPrevError = errors.New("transaction discarded because of previous errors")
	ErrMultiCommandNested        = errors.New("multi calls can not be nested")
	ErrDBIndexOutOfRange         = errors.New("(error) ERR DB index is out of range")
	ErrKeyNotFound               = errors.New("failed to find the key")
)
