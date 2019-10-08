package db

import (
	"database/sql"
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
		Ping() error
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
)

func (s *sqlDatabase) Ping() error {
	return s.db.Ping()
}
func (s *sqlDatabase) Query(query string, args ...interface{}) (rows, error) {
	return s.db.Query(query, args...)
}
func (s *sqlDatabase) QueryRow(query string, args ...interface{}) row {
	return s.db.QueryRow(query, args...)
}
func (s *sqlDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}
func (s *sqlDatabase) Begin() (transaction, error) {
	return s.db.Begin()
}