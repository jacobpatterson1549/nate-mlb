package db

import (
	"database/sql"
	"fmt"
	"time"
)

var (
	db *sql.DB
)

// InitDB initializes the pointer to the database
func InitDB(dataSourceName string) error {
	driverName := "postgres"
	var err error
	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("problem opening database %v", err)
	}
	return nil
}

func expectSingleRowAffected(r sql.Result) error {
	rows, err := r.RowsAffected()
	if err == nil && rows != 1 {
		return fmt.Errorf("expected to updated 1 row, but updated %d", rows)
	}
	return err
}

// GetUtcTime retrieves the current UTC time
func GetUtcTime() time.Time {
	return time.Now().UTC()
}
