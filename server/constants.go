package server

const (
	GET              string = "GET"
	SET              string = "SET"
	DEL              string = "DEL"
	INCR             string = "INCR"
	INCRBY           string = "INCRBY"
	MULTI            string = "MULTI"
	QUEUED           string = "QUEUED"
	EXEC             string = "EXEC"
	DISCARD          string = "DISCARD"
	COMPACT          string = "COMPACT"
	PING             string = "PING"
	PONG             string = "PONG"
	DISCONNECT       string = "DISCONNECT"
	SELECT           string = "SELECT"
	MSSG_EMPTY_ARRAY string = "(empty array)"
	MSSG_OK          string = "OK"
	MSSG_NIL         string = "(nil)"
	DB_RANGE_MIN     int    = 0
	DB_RANGE_MAX     int    = 15
)
