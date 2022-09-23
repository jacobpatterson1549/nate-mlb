package db

import (
	"database/sql"
	"fmt"
)

type (
	// sqlDatabase is a mockable database which conforms to the database interface
	// inspired from https://stackoverflow.com/questions/31364291/mocking-database-sql-structs-in-go
	// (https://github.com/EndFirstCorp/onedb)
	sqlDatabase struct {
		db *sql.DB
		database
	}

	database interface {
		Query(query string, args ...interface{}) (rows, error)
		QueryRow(query string, args ...interface{}) row
		Exec(query string, args ...interface{}) (sql.Result, error)
		Begin() (transaction, error)
	}
	row interface {
		Scan(dest ...interface{}) error
	}
	rows interface {
		Close() error
		Next() bool
		row // Scan method
	}
	transaction interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		Commit() error
		Rollback() error
	}

	sqlDB struct {
		db database
	}

	sqlTX struct {
		tx      transaction
		queries []writeSQLFunction
	}
)

func newSQLDatabase(driverName, dataSourceName string) (*sqlDB, error) {
	sqlDb, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening database %v", err)
	}
	d := sqlDB{
		db: sqlDatabase{
			db: sqlDb,
		},
	}
	return &d, nil
}

func (s sqlDatabase) Query(query string, args ...interface{}) (rows, error) {
	return s.db.Query(query, args...)
}
func (s sqlDatabase) QueryRow(query string, args ...interface{}) row {
	return s.db.QueryRow(query, args...)
}
func (s sqlDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}
func (s sqlDatabase) Begin() (transaction, error) {
	return s.db.Begin()
}

func (d *sqlDB) begin() (dbTX, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	t := sqlTX{
		tx: tx,
	}
	return &t, nil
}
