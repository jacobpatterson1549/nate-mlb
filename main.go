package main

import (
	"log"
	"os"
	"strconv"
)

func main() {

	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable not set")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	err = runServer(portNumber)
	if err != nil {
		log.Fatal(err)
	}
}
