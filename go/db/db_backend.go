package db

import (
	"context"
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
		ColumnTypes() ([]*sql.ColumnType, error)
		Columns() ([]string, error)
		Err() error
		Next() bool
		NextResultSet() bool
		row // Scan method
	}
	transaction interface {
		Commit() error
		Exec(query string, args ...interface{}) (sql.Result, error)
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		Prepare(query string) (*sql.Stmt, error)
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
		Rollback() error
		Stmt(stmt *sql.Stmt) *sql.Stmt
		StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	}
	// transaction interface {
	// 	Commit() error
	// 	Exec(query string, args ...interface{}) (sql.Result, error)
	// 	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// 	Prepare(query string) (stmt, error)
	// 	PrepareContext(ctx context.Context, query string) (stmt, error)
	// 	Query(query string, args ...interface{}) (rows, error)
	// 	QueryContext(ctx context.Context, query string, args ...interface{}) (rows, error)
	// 	QueryRow(query string, args ...interface{}) row
	// 	QueryRowContext(ctx context.Context, query string, args ...interface{}) row
	// 	Rollback() error
	// 	Stmt(stmt stmt) stmt
	// 	StmtContext(ctx context.Context, stmt stmt) stmt
	// }
	// stmt interface {
	// 	Close() error
	// 	Exec(args ...interface{}) (sql.Result, error)
	// 	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	// 	Query(args ...interface{}) (rows, error)
	// 	QueryContext(ctx context.Context, args ...interface{}) (rows, error)
	// 	QueryRow(args ...interface{}) row
	// 	QueryRowContext(ctx context.Context, args ...interface{}) row
	// }
	// columnType interface {
	// 	DatabaseTypeName() string
	// 	DecimalSize() (precision, scale int64, ok bool)
	// 	Length() (length int64, ok bool)
	// 	Name() string
	// 	Nullable() (nullable, ok bool)
	// 	ScanType() reflect.Type
	// }
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
