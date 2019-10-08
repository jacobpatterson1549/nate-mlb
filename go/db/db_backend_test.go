package db

import (
	"database/sql"
)

type (
	// mockDatabase implements the database interface
	mockDatabase struct {
		PingFunc     func() error
		QueryFunc    func(query string, args ...interface{}) (rows, error)
		QueryRowFunc func(query string, args ...interface{}) row
		ExecFunc     func(query string, args ...interface{}) (sql.Result, error)
		BeginFunc    func() (transaction, error)
	}
	// mockRow implements the row interface
	mockRow struct {
		ScanFunc func(dest ...interface{}) error
	}
	// mockRows implements the rows interface
	mockRows struct {
		CloseFunc func() error
		NextFunc  func() bool
		row       // Scan method
	}
	// mockTransaction implements the transaction interface
	mockTransaction struct {
		ExecFunc     func(query string, args ...interface{}) (sql.Result, error)
		CommitFunc   func() error
		RollbackFunc func() error
	}
)

func (m *mockDatabase) Ping() error {
	return m.PingFunc()
}
func (m *mockDatabase) Query(query string, args ...interface{}) (rows, error) {
	return m.QueryFunc(query, args)
}
func (m *mockDatabase) QueryRow(query string, args ...interface{}) row {
	return m.QueryRowFunc(query, args)
}
func (m *mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args)
}
func (m *mockDatabase) Begin() (transaction, error) {
	return m.BeginFunc()
}

func (m *mockRow) Scan(dest ...interface{}) error {
	return m.ScanFunc(dest)
}

func (m *mockRows) Close() error {
	return m.CloseFunc()
}
func (m *mockRows) Next() bool {
	return m.NextFunc()
}

func (m *mockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args)
}
func (m *mockTransaction) Commit() error {
	return m.CommitFunc()
}
func (m *mockTransaction) Rollback() error {
	return m.RollbackFunc()
}
