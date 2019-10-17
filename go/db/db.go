// Package db contains persistence functions to store and query data for the server.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type (
	// ID is used to identify an item in the database or a relation to another noun's id
	ID int

	readSQLFunction struct {
		name string
		cols []string
		args []interface{}
	}

	writeSQLFunction struct {
		name string
		args []interface{}
	}
)

// Datastore interface can be used to access and persist data
type Datastore struct {
	db           database
	ph           passwordHasher
	sportTypes   SportTypeMap
	playerTypes  PlayerTypeMap
	readFileFunc func(filename string) ([]byte, error)
	readDirFunc  func(dirname string) ([]os.FileInfo, error)
}

// NewDatastore creates a new sqlDatastore
func NewDatastore(dataSourceName string) (*Datastore, error) {
	db, err := newSQLDatabase(dataSourceName)
	if err != nil {
		return nil, err
	}

	sleepFunc := func(sleepSeconds int) {
		s := fmt.Sprintf("%ds", sleepSeconds)
		d, err := time.ParseDuration(s)
		if err != nil {
			panic(err)
		}
		time.Sleep(d) // BLOCKING
	}
	numFibonacciTries := 7
	if err := waitForDb(db, sleepFunc, numFibonacciTries); err != nil {
		return nil, fmt.Errorf("establishing connection: %v", err)
	}

	ds := Datastore{
		db:           db,
		ph:           bcryptPasswordHasher{},
		readFileFunc: ioutil.ReadFile,
		readDirFunc:  ioutil.ReadDir,
	}

	if err = ds.SetupTablesAndFunctions(); err != nil {
		return nil, err
	}

	sportTypes, err := ds.GetSportTypes()
	if err != nil {
		return nil, err
	}
	ds.sportTypes = sportTypes

	playerTypes, err := ds.GetPlayerTypes()
	if err != nil {
		return nil, err
	}
	ds.playerTypes = playerTypes

	return &ds, nil
}

// Ping ensures the database connection is active and returns an error if not
func (ds Datastore) Ping() error {
	return ds.db.Ping()
}

// GetUtcTime retrieves the current UTC time
func (Datastore) GetUtcTime() time.Time {
	return time.Now().UTC()
}

// SportTypes implements the SportTypeGetter interface for Datastore
func (ds Datastore) SportTypes() SportTypeMap {
	return ds.sportTypes
}

// PlayerTypes implements the PlayerTypeGetter interface for Datastore
func (ds Datastore) PlayerTypes() PlayerTypeMap {
	return ds.playerTypes
}

func (ds Datastore) executeInTransaction(queries []writeSQLFunction) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}
	var result sql.Result
	for _, sqlFunction := range queries {
		result, err = tx.Exec(sqlFunction.sql(), sqlFunction.args...)
		if err == nil {
			err = expectSingleRowAffected(result)
		}
		if err != nil {
			err = fmt.Errorf("%s: %w", sqlFunction.name, err)
			break
		}
	}
	switch {
	case err != nil:
		err = fmt.Errorf("saving: %w", err)
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			err = fmt.Errorf("%w ROLLBACK ERROR: %w", err, rollbackErr)
		}
	case len(queries) > 0:
		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("committing transaction to save: %w", err)
		}
	}
	return err
}

func expectSingleRowAffected(r sql.Result) error {
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to update 1 row, but updated %d", rows)
	}
	return nil
}

func expectRowFound(row row) error {
	var found bool
	err := row.Scan(&found)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("expected to update at least one row, but did not")
	}
	return nil
}

func newReadSQLFunction(name string, cols []string, args ...interface{}) readSQLFunction {
	return readSQLFunction{
		name: name,
		cols: cols,
		args: args,
	}
}

func newWriteSQLFunction(name string, args ...interface{}) writeSQLFunction {
	return writeSQLFunction{
		name: name,
		args: args,
	}
}

func (f readSQLFunction) sql() string {
	argIndexes := make([]string, len(f.args))
	for i := range argIndexes {
		argIndexes[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("SELECT %s FROM %s(%s)", strings.Join(f.cols, ", "), f.name, strings.Join(argIndexes, ", "))
}

func (f writeSQLFunction) sql() string {
	argIndexes := make([]string, len(f.args))
	for i := range argIndexes {
		argIndexes[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("SELECT %s(%s)", f.name, strings.Join(argIndexes, ", "))
}

// waitForDb tries to ensure the database connection is valid, waiting a fibonacci amount of seconds between attempts
func waitForDb(d database, sleepFunc func(sleepSeconds int), numFibonacciTries int) error {
	a, b := 1, 0
	var err error
	for i := 0; i < numFibonacciTries; i++ {
		err = d.Ping()
		if err == nil {
			log.Println("connected to database")
			return nil
		}
		log.Printf("failed to connect to database; trying again in %v seconds...\n", b)
		sleepFunc(b)
		c := b
		b = a
		a = b + c
	}
	return err
}
