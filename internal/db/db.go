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
	args []interface{}
}

// InitDB initializes the pointer to the database
func InitDB(driverName, dataSourceName string) error {
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

func exececuteInTransaction(queries []query) error {
	var err error
	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("problem starting transaction to save: %v", err)
	}
	var result sql.Result
	for _, query := range queries {
		if err == nil {
			result, err = tx.Exec(query.sql, query.args...)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	if err == nil {
		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("problem committing transaction to save: %v", err)
		}
	} else {
		err = fmt.Errorf("problem saving: %v", err)
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("%v, ROLLBACK ERROR: %v", err, err2)
		}
	}
	return err
}

func expectSingleRowAffected(r sql.Result) error {
	rows, err := r.RowsAffected()
	if err == nil && rows != 1 {
		err = fmt.Errorf("expected to update 1 row, but updated %d", rows)
	}
	return err
}
