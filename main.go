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
	err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
    }

	port, ok := os.LookupEnv("PORT")
	if !ok {
		fmt.Println("Couldn't find the port variable")
	}

	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Error while setting up listener: %v", err)
	}

	s := &server.Server{
		Db:  db.GetNewDB(inMemoryStore.NewInMemoryStore()),
		Listener: ln,
	}

	s.Start();

}
