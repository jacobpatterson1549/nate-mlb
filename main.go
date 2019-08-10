package main

import (
	"log"
	"os"
	"strconv"
)

func main() {

	err := InitDB()
	if err != nil {
		log.Fatal("Could not set database ", err)
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable not set")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("PORT (%v) invalid as number: %v", port, err)
	}

	err = runServer(portNumber)
	if err != nil {
		log.Fatal(err)
	}
}
