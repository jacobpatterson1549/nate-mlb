// Package db contains persistence functions to store and query data for the server.
package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
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

	datastoreConfig struct {
		driverName           string
		dataSourceName       string
		ph                   passwordHasher
		pingFailureSleepFunc func(sleepSeconds int)
		log                  *log.Logger
		fs                   fs.ReadFileFS
	}

	// Datastore interface can be used to access and persist data
	Datastore struct {
		db          database
		fs          fs.ReadFileFS
		ph          passwordHasher
		sportTypes  SportTypeMap
		playerTypes PlayerTypeMap
		log         *log.Logger
	}
)

// NewDatastore creates a new sqlDatastore
func NewDatastore(dataSourceName string, log *log.Logger, fs fs.ReadFileFS) (*Datastore, error) {
	cfg := datastoreConfig{
		driverName:     "postgres",
		dataSourceName: dataSourceName,
		ph:             bcryptPasswordHasher{},
		pingFailureSleepFunc: func(sleepSeconds int) {
			s := fmt.Sprintf("%ds", sleepSeconds)
			d, err := time.ParseDuration(s)
			if err != nil {
				panic(err)
			}
			time.Sleep(d) // BLOCKING
		},
		log: log,
		fs:  fs,
	}
	return newDatastore(cfg)
}

func newDatastore(cfg datastoreConfig) (*Datastore, error) {
	db, err := newSQLDatabase(cfg.driverName, cfg.dataSourceName)
	if err != nil {
		return nil, err
	}

	ds := Datastore{
		db:  db,
		fs:  cfg.fs,
		ph:  cfg.ph,
		log: cfg.log,
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
			err = fmt.Errorf("%v and ROLLBACK ERROR: %w", err, rollbackErr)
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
