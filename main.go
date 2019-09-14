package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/server"
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
	adminPassword      string
)

func init() {
	databaseDriverName = "postgres"
	flag.StringVar(&dataSourceName, "ds", "", "The data source to the PostgreSQL database (connection URI).  Defaults to "+environmentVariableDatabaseURL)
	flag.IntVar(&portNumber, "p", 0, "The port number to run the server on.  Defaults to "+environmentVariablePort)
	flag.StringVar(&adminPassword, "ap", "", "The admin user password.  Requires the -ds option.")

	fmt.Println(databaseDriverName, dataSourceName, portNumber, adminPassword)
	var ok bool
	if len(dataSourceName) == 0 {
		dataSourceName, ok = os.LookupEnv(environmentVariableDatabaseURL)
		if !ok {
			log.Fatal(environmentVariableDatabaseURL, " environment variable not set")
		}
	}
	if len(adminPassword) == 0 && portNumber == 0 {
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
}

func main() {
	var err error
	switch {
	case len(adminPassword) != 0:
		err = db.SetAdminPassword(adminPassword)
	default:
		err = server.Run(portNumber, databaseDriverName, dataSourceName)
	}
	if err != nil {
		log.Fatal(err)
	}
}
