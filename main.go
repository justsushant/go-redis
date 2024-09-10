package main

import (
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/db"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/server"
	"github.com/justsushant/one2n-go-bootcamp/go-redis/store/inMemoryStore"
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
	s := &server.Server{
		Db:       map[int]db.DbInterface{0: db.GetNewDB(inMemoryStore.NewInMemoryStore())},
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