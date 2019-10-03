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

type mainVars struct {
	adminPassword   string
	applicationName string
	dataSourceName  string
	port            string
	playerTypesCsv  string
}

func main() {
	mainVars := initFlags()
	for _, startupFunc := range startupFuncs(mainVars) {
		if err := startupFunc(); err != nil {
			log.Fatal(err)
		}
	}
}

func usage(programName string) {
	envVars := []string{
		environmentVariableDatabaseURL,
		environmentVariablePort,
		environmentVariableApplicationName,
		environmentVariableAdminPassword,
	}
	fmt.Fprintln(flag.CommandLine.Output(), "Starts the server")
	fmt.Fprintln(flag.CommandLine.Output(), "Reads environment variables when possible:", fmt.Sprintf("[%s]", strings.Join(envVars, ",")))
	fmt.Fprintln(flag.CommandLine.Output(), fmt.Sprintf("Usage of %s:", programName))
	flag.PrintDefaults()
}

func initFlags() mainVars {
	programName := os.Args[0]
	flag.Usage = func() { usage(programName) }
	mainVars := mainVars{}
	defaultApplicationName := func() string {
		applicationName, ok := os.LookupEnv(environmentVariableApplicationName)
		if !ok {
			return programName
		}
		return applicationName
	}
	flag.StringVar(&mainVars.adminPassword, "ap", os.Getenv(environmentVariableAdminPassword), "The admin user password to set.")
	flag.StringVar(&mainVars.applicationName, "n", defaultApplicationName(), "The name of the application.")
	flag.StringVar(&mainVars.dataSourceName, "ds", os.Getenv(environmentVariableDatabaseURL), "The data source to the PostgreSQL database (connection URI).")
	flag.StringVar(&mainVars.port, "p", os.Getenv(environmentVariablePort), "The port number to run the server on.")
	flag.StringVar(&mainVars.playerTypesCsv, "pt", os.Getenv(environmentVariablePlayerTypesCsv), "A csv whitelist of player types to use. Must not contain spaces.")
	flag.Parse()
	return mainVars
}

func startupFuncs(mainVars mainVars) []func() error {
	startupFuncs := make([]func() error, 0, 6)
	startupFuncs = append(startupFuncs, func() error { return db.Init(mainVars.dataSourceName) })
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
	if len(mainVars.playerTypesCsv) != 0 {
		startupFuncs = append(startupFuncs, func() error { return db.LimitPlayerTypes(mainVars.playerTypesCsv) })
	}
	if len(mainVars.adminPassword) != 0 {
		startupFuncs = append(startupFuncs, func() error { return db.SetAdminPassword(mainVars.adminPassword) })
	}
	return append(startupFuncs, func() error { return server.Run(mainVars.port, mainVars.applicationName) })
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
