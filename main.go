// Package main can run the server from the main() function.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/server"
	_ "github.com/lib/pq"
)

const (
	environmentVariableAdminPassword   = "ADMIN_PASSWORD"
	environmentVariableApplicationName = "APPLICATION_NAME"
	environmentVariableDatabaseURL     = "DATABASE_URL"
	environmentVariablePort            = "PORT"
	environmentVariablePlayerTypesCsv  = "PLAYER_TYPES"
)

var (
	adminPassword   string
	applicationName string
	dataSourceName  string
	port            string
	playerTypesCsv  string
)

func usage() {
	envVars := []string{
		environmentVariableDatabaseURL,
		environmentVariablePort,
		environmentVariableApplicationName,
		environmentVariableAdminPassword,
	}
	fmt.Fprintln(flag.CommandLine.Output(), "Starts the server")
	fmt.Fprintln(flag.CommandLine.Output(), "Reads environment variables when possible:", fmt.Sprintf("[%s]", strings.Join(envVars, ",")))
	fmt.Fprintln(flag.CommandLine.Output(), "Usage of", os.Args[0], ":")
	flag.PrintDefaults()
}

func initFlags() {
	flag.Usage = usage
	defaultApplicationName := func() string {
		applicationName, ok := os.LookupEnv(environmentVariableApplicationName)
		if !ok {
			return os.Args[0]
		}
		return applicationName
	}
	flag.StringVar(&adminPassword, "ap", os.Getenv(environmentVariableAdminPassword), "The admin user password to set.")
	flag.StringVar(&applicationName, "n", defaultApplicationName(), "The name of the application.")
	flag.StringVar(&dataSourceName, "ds", os.Getenv(environmentVariableDatabaseURL), "The data source to the PostgreSQL database (connection URI).")
	flag.StringVar(&port, "p", os.Getenv(environmentVariablePort), "The port number to run the server on.")
	flag.StringVar(&playerTypesCsv, "pt", os.Getenv(environmentVariablePlayerTypesCsv), "A csv whitelist of player types to use. Must not contain spaces.")
	flag.Parse()
}

func startupFuncs() []func() error {
	startupFuncs := make([]func() error, 0, 6)
	startupFuncs = append(startupFuncs, func() error { return db.Init(dataSourceName) })
	startupFuncs = append(startupFuncs, func() error {
		sleepFunc := func(sleepSeconds int) {
			s := fmt.Sprintf("%ds", sleepSeconds)
			d, err := time.ParseDuration(s)
			if err != nil {
				panic(err)
			}
			time.Sleep(d) // BLOCKING
		}
		return waitForDb(db.Ping, sleepFunc, 7)
	})
	startupFuncs = append(startupFuncs, db.SetupTablesAndFunctions)
	startupFuncs = append(startupFuncs, db.LoadSportTypes)
	startupFuncs = append(startupFuncs, db.LoadPlayerTypes)
	if len(playerTypesCsv) != 0 {
		startupFuncs = append(startupFuncs, func() error { return db.LimitPlayerTypes(playerTypesCsv) })
	}
	if len(adminPassword) != 0 {
		startupFuncs = append(startupFuncs, func() error { return db.SetAdminPassword(adminPassword) })
	}
	return append(startupFuncs, func() error { return server.Run(port, applicationName) })
}

func main() {
	initFlags()
	startupFuncs := startupFuncs()
	for _, startupFunc := range startupFuncs {
		err := startupFunc()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// waitForDb tries to ensure the database connection is valid, waiting a fibonacci amount of seconds between attempts
func waitForDb(dbCheckFunc func() error, sleepFunc func(sleepSeconds int), numFibonacciTries int) error {
	a, b := 1, 0
	var err error
	for i := 0; i < numFibonacciTries; i++ {
		err = dbCheckFunc()
		if err == nil {
			fmt.Println("connected to database")
			return nil
		}
		fmt.Printf("failed to connect to database; trying again in %v seconds...\n", b)
		sleepFunc(b)
		c := b
		b = a
		a = b + c
	}
	return err
}
