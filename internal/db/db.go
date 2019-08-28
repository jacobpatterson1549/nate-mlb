package db

import (
	"database/sql"
	"fmt"
	"time"
)

var (
	db *sql.DB
)

type query struct {
	sql  string
	args []interface{} // TODO: make this ...interface{}
}

// Init initializes the pointer to the database
func Init(driverName, dataSourceName string) error {
	var err error
	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("problem opening database %v", err)
	}
	return nil
}

// GetUtcTime retrieves the current UTC time
func GetUtcTime() time.Time {
	return time.Now().UTC()
}

func exececuteInTransaction(queries <-chan query, quit chan<- error) {
	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("problem starting transaction to save: %v", err)
	}
	var result sql.Result
	for query := range queries {
		result, err = tx.Exec(query.sql, query.args...)
		if err != nil {
			break
		}
		err = expectSingleRowAffected(result)
		if err != nil {
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("problem saving: %v", err)
		if err2 := tx.Rollback(); err2 != nil { // TODO: rename to errRollback
			err = fmt.Errorf("%v, ROLLBACK ERROR: %v", err, err2)
		}
	} else {
		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("problem committing transaction to save: %v", err)
		}
	}
	quit <- err
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
