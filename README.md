# Key-Value DB (Redis) Exercise

This exercise involves creating a simplified clone of the popular in-memory key-value store, Redis. It is designed to emulate basic functionalities of Redis, including data storage, retrieval, and manipulation, with support for transactions and database indexing.

The server can be started to accept connections via an HTTP client, allowing interaction over HTTP requests.

## Supported Commands

The commands work similarly to those in actual Redis. The list of supported commands is as follows:
- **GET**: retrieves a record
- **SET**: sets a record
- **DEL**: deletes a record
- **INCR**: increments an integer value by 1
- **INCRBY**: increments an integer value by the specified number
- **MULTI**: initiates a transaction
- **EXEC**: executes a transaction
- **DISCARD**: discards a transaction
- **COMPACT**: returns the current state of the store
- **DISCONNECT**: disconnects the client

## Usage 

1. Run the command below to build and run the binary:
   ```
   make run
   ```
2. You can connect to this server using any HTTP client, say netcat (assuming, server is running on localhost:8080)
```
nc localhost 8080
```

## Improvements
- Write test for disconnection of the server
- Improve the error message for invalid number of arguments to a command
- Add comments for db package and improve the tests
- Look for more features to clone


