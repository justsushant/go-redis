package main

import (
	"net"
	"os"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/server"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

func main() {
	// load env file
	err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
    }

	// set the port
	port, ok := os.LookupEnv("PORT")
	if !ok {
		fmt.Println("Couldn't find the port variable")
	}

	// start listening
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Error while setting up listener: %v", err)
	}

	// create a new server
	s := &server.Server{
		Db:  map[int]db.DbInterface{0:db.GetNewDB(inMemoryStore.NewInMemoryStore())},
		Listener: ln,
	}

	// start the server
	s.Start();

}
