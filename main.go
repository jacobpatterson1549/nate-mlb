package main

import (
	"log"
	"os"
	"strconv"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port not set.")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	startServer(portNumber)
}
