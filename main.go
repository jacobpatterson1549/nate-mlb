package main

import (
	"log"
	"nate-mlb/internal/server"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	environmentVariableDatabaseURL = "DATABASE_URL"
	environmentVariablePort        = "PORT"
)

var (
	databaseDriverName string
	dataSourceName     string
	portNumber         int
)

func init() {
	var ok bool

	databaseDriverName = "postgres"

	dataSourceName, ok = os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal(environmentVariableDatabaseURL, " environment variable not set")
	}

	var port string
	port, ok = os.LookupEnv(environmentVariablePort)
	if !ok {
		log.Fatal(environmentVariablePort, " environment variable not set")
	}
	var err error
	portNumber, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(environmentVariablePort, " environment variable is invalid: ", port)
	}
}

func main() {
	err := server.Run(portNumber, databaseDriverName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
}
