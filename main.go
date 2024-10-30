package main

import (
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/db"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/db/keyvaldb"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/server"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/store/inmemorystore"
)

const DEFAULT_PORT = "8080"

func main() {
	// load env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// set the port
	port := getEnv("PORT", DEFAULT_PORT)

	// start listening
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Error while setting up listener: %v", err)
	}

	// create a new server
	// TODO: Fix the name of constructors to somethig more suitable according to context
	// TODO: FIx the name of interfaces according to their behaviour, usually end with "er"
	// TODO: Encapsulate it in a constructor
	s := &server.Server{
		Db:       map[int]db.Database{0: keyvaldb.GetNewDB(inmemorystore.NewInMemoryStore())},
		Listener: ln,
	}

	// start the server
	s.Start()
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}
