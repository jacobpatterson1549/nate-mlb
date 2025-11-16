// Package main can run the server from the main() function.
package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
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
	environmentVariableNflAppKey       = "NFL_APP_KEY"
	environmentVariableLogRequestURIs  = "LOG_REQUEST_URIS"
)

var (
	//go:embed sql
	sqlFS embed.FS
	//go:embed html
	htmlFS embed.FS
	//go:embed js
	jsFS embed.FS
	//go:embed static
	staticFS embed.FS
)

type mainFlags struct {
	adminPassword   string
	applicationName string
	dataSourceName  string
	port            string
	playerTypesCsv  string
	nflAppKey       string
	logRequestURIs  bool
}

func main() {
	fs, mainFlags := initFlags(os.Args[0])
	flag.CommandLine = fs
	flag.Parse()

	var buf bytes.Buffer
	log := log.New(&buf, mainFlags.applicationName, log.LstdFlags)
	log.SetOutput(os.Stdout)

	startupFuncs := startupFuncs(mainFlags, log)
	for _, startupFunc := range startupFuncs {
		if err := startupFunc(); err != nil {
			log.Fatal(err)
		}
	}
}

func flagUsage(fs *flag.FlagSet) {
	envVars := []string{
		environmentVariableDatabaseURL,
		environmentVariablePort,
		environmentVariableApplicationName,
		environmentVariableAdminPassword,
		environmentVariablePlayerTypesCsv,
		environmentVariableNflAppKey,
	}
	fmt.Fprintln(fs.Output(), "Starts the server")
	fmt.Fprintln(fs.Output(), "Reads environment variables when possible:", fmt.Sprintf("[%s]", strings.Join(envVars, ",")))
	fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
	fs.PrintDefaults()
}

func initFlags(programName string) (*flag.FlagSet, *mainFlags) {
	fs := flag.NewFlagSet(programName, flag.ExitOnError)
	fs.Usage = func() { flagUsage(fs) }
	mainFlags := new(mainFlags)
	defaultApplicationName := func() string {
		if applicationName, ok := os.LookupEnv(environmentVariableApplicationName); ok {
			return applicationName
		}
		return programName
	}
	fs.StringVar(&mainFlags.adminPassword, "ap", os.Getenv(environmentVariableAdminPassword), "The admin user password to set.")
	fs.StringVar(&mainFlags.applicationName, "n", defaultApplicationName(), "The name of the application.  Also used as the deploy environment name.")
	fs.StringVar(&mainFlags.dataSourceName, "ds", os.Getenv(environmentVariableDatabaseURL), "The data source to the PostgreSQL database (connection URI).")
	fs.StringVar(&mainFlags.port, "p", os.Getenv(environmentVariablePort), "The port number to run the server on.")
	fs.StringVar(&mainFlags.playerTypesCsv, "pt", os.Getenv(environmentVariablePlayerTypesCsv), "A csv whitelist of player types to use. Must not contain spaces.")
	fs.StringVar(&mainFlags.nflAppKey, "ak", os.Getenv(environmentVariableNflAppKey), "The application key used to make nfl requests")
	_, logRequestURIs := os.LookupEnv(environmentVariableLogRequestURIs)
	fs.BoolVar(&mainFlags.logRequestURIs, "logRequestURIs", logRequestURIs, "logs the uris of requests to external sources for data when set")
	return fs, mainFlags
}

func startupFuncs(mainFlags *mainFlags, log *log.Logger) []func() error {
	var ds *db.Datastore
	startupFuncs := make([]func() error, 0, 2)
	startupFuncs = append(startupFuncs, func() error {
		var err error
		ds, err = db.NewDatastore(mainFlags.dataSourceName, log, sqlFS)
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
		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		cfg := server.Config{
			DisplayName:  mainFlags.applicationName,
			NflAppKey:    mainFlags.nflAppKey,
			Port:         mainFlags.port,
			HTMLFS:       htmlFS,
			JavascriptFS: jsFS,
			StaticFS:     staticFS,
		}
		server, err := cfg.New(log, ds, httpClient)
		if err != nil {
			return err
		}
		return server.Run()
	})
}
