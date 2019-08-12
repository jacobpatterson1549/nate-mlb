package main

import (
	"log"
	"nate-mlb/internal/db"
	"nate-mlb/internal/server"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {

	dataSourceName, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable not set")
	}

	err := db.InitDB(dataSourceName)
	if err != nil {
		log.Fatal("Could not set database ", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("PORT (%v) invalid as number: %v", port, err)
	}

	err = server.Run(portNumber)
	if err != nil {
		log.Fatal(err)
	}
}
