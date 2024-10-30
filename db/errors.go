package db

import "errors"

var (
	ErrKeyNotFound   = errors.New("(nil)")
	ErrKeyNotInteger = errors.New("value is not an integer or out of range")
)
