package main

import (
	"os"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/server"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/db"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
)

func main() {
	s := &server.Server{
		Db:  db.GetNewDB(inMemoryStore.NewInMemoryStore()),
		Out: os.Stdout,
	}

	s.Start();

}
