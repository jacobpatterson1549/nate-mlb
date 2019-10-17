// Package main can run the server from the main() function.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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

type mainFlags struct {
	adminPassword   string
	applicationName string
	dataSourceName  string
	port            string
	playerTypesCsv  string
}

var ds *db.Datastore

func main() {
	mainFlags := initFlags()
	for _, startupFunc := range startupFuncs(mainFlags) {
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

func initFlags() mainFlags {
	programName := os.Args[0]
	flag.Usage = func() { usage(programName) }
	mainFlags := mainFlags{}
	defaultApplicationName := func() string {
		if applicationName, ok := os.LookupEnv(environmentVariableApplicationName); ok {
			return applicationName
		}
		return programName
	}
	flag.StringVar(&mainFlags.adminPassword, "ap", os.Getenv(environmentVariableAdminPassword), "The admin user password to set.")
	flag.StringVar(&mainFlags.applicationName, "n", defaultApplicationName(), "The name of the application.")
	flag.StringVar(&mainFlags.dataSourceName, "ds", os.Getenv(environmentVariableDatabaseURL), "The data source to the PostgreSQL database (connection URI).")
	flag.StringVar(&mainFlags.port, "p", os.Getenv(environmentVariablePort), "The port number to run the server on.")
	flag.StringVar(&mainFlags.playerTypesCsv, "pt", os.Getenv(environmentVariablePlayerTypesCsv), "A csv whitelist of player types to use. Must not contain spaces.")
	flag.Parse()
	return mainFlags
}

func startupFuncs(mainFlags mainFlags) []func() error {
	startupFuncs := make([]func() error, 0, 2)
	startupFuncs = append(startupFuncs, func() error {
		var err error
		ds, err = db.NewDatastore(mainFlags.dataSourceName)
		return err
	})
	if len(mainFlags.playerTypesCsv) != 0 {
		startupFuncs = append(startupFuncs, func() error {
			return ds.LimitPlayerTypes(mainFlags.playerTypesCsv)
		})
	}
	if len(mainFlags.adminPassword) != 0 {
		startupFuncs = append(startupFuncs, func() error {
			return ds.SetAdminPassword(db.Password(mainFlags.adminPassword))
		})
	}
	return append(startupFuncs, func() error {
		cfg, err := server.NewConfig(mainFlags.applicationName, ds, mainFlags.port)
		if err != nil {
			return err
		}
		return server.Run(*cfg)
	})
}
