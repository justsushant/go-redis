
# go-redis

This project is a simplified clone of the popular in-memory data structure store, Redis. It is designed to emulate basic functionalities of Redis, including data storage, retrieval, and manipulation, with support for transactions and database index.

The server can be started to accept connections via an HTTP client, allowing interaction over HTTP requests.

## Supported Commands
- GET key: retrives record
- SET key val: sets record
- DEL key: deletes record
- INCR key: increments integer value by 1
- INCRBY key val: increments integer value by specified number
- MULTI: initiates transaction
- EXEC: executes transaction
- DISCARD: discards transaction
- COMPACT: returns the current state of store
- DISCONNECT: disconnects the client

## Run 
This requires Go installation to generate the binary for go-redis server. 
```
go build -o go-redis .
./go-redis
```

You can connect to this server using any HTTP client, say netcat (assuming, server is running on localhost:8080)
```
nc localhost 8080
```

## Improvements
- Write test for disconnection of the server
- Improve the error message for invalid number of arguments to a command
- Add comments for db package and improve the tests
- Look for more features to clone


