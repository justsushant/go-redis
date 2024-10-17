package main

// run the app using make run
// load 50k key-val pairs in the go-redis
// then do the following:
// 1. time to fetch an already existing key
// 2. time to set a new key
// 3. time to get a newly set key

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	BatchSize  = 10000
	ServerAddr = "localhost:8080"
)

func main() {
	var wg sync.WaitGroup

	// dir, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println(dir)

	// run the app
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	cmd := exec.Command("make", "run")
	// 	output, err := cmd.CombinedOutput()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	log.Println("app running:", output)
	// }()
	// wg.Wait()

	// loading 50k key-val pairs into go-redis
	startTimeToLoad := time.Now()
	for i := 0; i < 5; i++ {
		numStart := i * BatchSize
		numEnd := (i+1)*BatchSize - 1

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := numStart; i <= numEnd; i++ {
				// key := "key" + string(i)
				key := fmt.Sprintf("key%d", i)
				// val := string(i)
				val := fmt.Sprintf("%d", i)

				err := setKeyAndVal(key, val)
				if err != nil {
					log.Printf("error at number %d:%v\n", i, err)
				}
			}
		}()
	}
	wg.Wait()
	timeElapsedToLoad := time.Since(startTimeToLoad)
	log.Println("Time taken to load 50k key-val pairs: ", timeElapsedToLoad)

	// time taken to get a newly set key
	id := BatchSize*3 - 1
	existingKey := fmt.Sprintf("key%d", id)
	existingVal := fmt.Sprintf("\"%d\"", id)
	timeElapsedToGetExistingKey := calcTimeForGetKey(existingKey, existingVal)
	log.Printf("Time taken to get an existing key-val pair (id %d): %v\n", id, timeElapsedToGetExistingKey)

	// time taken to set a key
	timeElapsedToSet := calcTimeForSetKey("load", "test")
	log.Println("Time taken to set a key-val pair: ", timeElapsedToSet)

	// time taken to get a newly set key
	timeElapsedToGet := calcTimeForGetKey("load", "\"test\"")
	log.Println("Time taken to get a key-val pair: ", timeElapsedToGet)
}

func calcTimeForSetKey(key, val string) time.Duration {
	start := time.Now()

	err := setKeyAndVal(key, val)
	if err != nil {
		log.Println("error while calculating time for set key:", err)
	}

	return time.Since(start)
}

func calcTimeForGetKey(key, val string) time.Duration {
	start := time.Now()

	err := getKey(key, val)
	if err != nil {
		log.Println("error while calculating time for set key:", err)
	}

	return time.Since(start)
}

func setKeyAndVal(key, val string) error {
	buff := make([]byte, 1024)
	data := []byte(fmt.Sprintf("SET %s %s", key, val))

	conn, err := net.Dial("tcp", ServerAddr)
	if err != nil {
		return fmt.Errorf("error while establishing conn: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error while writing to conn object: %v", err)
	}

	n, err := conn.Read(buff)
	if err != nil {
		return fmt.Errorf("error while reading from conn object: %v", err)
	}

	if string(buff[:n]) != "OK\n" {
		return fmt.Errorf("expected %q but got %q", "OK", string(buff[:n]))
	}

	return nil
}

func getKey(key, val string) error {
	buff := make([]byte, 1024)
	data := []byte(fmt.Sprintf("GET %s", key))

	conn, err := net.Dial("tcp", ServerAddr)
	if err != nil {
		return fmt.Errorf("error while establishing conn: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error while writing to conn object: %v", err)
	}

	n, err := conn.Read(buff)
	if err != nil {
		return fmt.Errorf("error while reading from conn object: %v", err)
	}

	if strings.TrimSpace(string(buff[:n])) != val {
		return fmt.Errorf("expected %q but got %q", val, strings.TrimSpace(string(buff[:n])))
	}

	return nil
}
