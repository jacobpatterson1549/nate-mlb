// Package main can run the server or set the admin password from the main() function.
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
	environmentVariableDatabaseURL     = "DATABASE_URL"
	environmentVariablePort            = "PORT"
	environmentVariableApplicationName = "APPLICATION_NAME"
	environmentVariableAdminPassword   = "ADMIN_PASSWORD"
)

var (
	databaseDriverName string
	dataSourceName     string
	port               string
	adminPassword      string
	applicationName    string
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

func init() {
	flag.Usage = usage
	databaseDriverName = "postgres"
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
	flag.Parse()
}

func main() {
	startupFuncs := make([]func() error, 0, 3)
	startupFuncs = append(startupFuncs, func() error { return db.Init(databaseDriverName, dataSourceName) })
	startupFuncs = append(startupFuncs, func() error { return db.SetupTablesAndFunctions() })
	if len(adminPassword) != 0 {
		startupFuncs = append(startupFuncs, func() error { return db.SetAdminPassword(adminPassword) })
	}
	startupFuncs = append(startupFuncs, func() error { return server.Run(port, applicationName) })

	for _, startupFunc := range startupFuncs {
		err := startupFunc()
		if err != nil {
			log.Fatal(err)
		}
	}
}
