package db

import "errors"

var (
	ErrKeyNotFound   = errors.New("(nil)")
	ErrKeyNotInteger = errors.New("value is not an integer or out of range")
)

const (
	SET_SUCCESS_MESSAGE    = "OK"
	DELETE_SUCCESS_MESSAGE = "(integer) 1"
	DELETE_FAILED_MESSAGE  = "(integer) 0"
	DEFAULT_INTEGER_VALUE  = "1"
	INTEGER                = "(integer)"
)
