// Package db contains persistence functions to store and query data for the server.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var db database

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

// Init initializes the pointer to the database
func Init(dataSourceName string) error {
	sqlDb, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("opening database %v", err)
	}
	db = &sqlDatabase{db: sqlDb}
	ph = bcryptPasswordHasher{}
	return nil
}

// Ping ensures the database connection is active and returns an error if not
func Ping() error {
	return db.Ping()
}

// GetUtcTime retrieves the current UTC time
func GetUtcTime() time.Time {
	return time.Now().UTC()
}

func executeInTransaction(queries []writeSQLFunction) error {
	tx, err := db.Begin()
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
