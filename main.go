package main

import (
	"log"
	"nate-mlb/internal/server"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

// names of variables which should be specied in the runtime environment or be in the .env file
const (
	EnvironmentVariableDatabaseURL = "DATABASE_URL"
	EnvironmentVariablePort        = "PORT"
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
		log.Fatal(EnvironmentVariableDatabaseURL, " environment variable not set")
	}

	var port string
	port, ok = os.LookupEnv(EnvironmentVariablePort)
	if !ok {
		log.Fatal(EnvironmentVariablePort, " environment variable not set")
	}
	var err error
	portNumber, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(EnvironmentVariablePort, " environment variable is invalid: ", port)
	}
}

func main() {
	err := server.Run(portNumber, databaseDriverName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
}
