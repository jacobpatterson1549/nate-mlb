package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
)

type (
	mockDriver struct {
		OpenFunc func(name string) (driver.Conn, error)
	}
	mockConn struct {
		closed      bool
		PrepareFunc func(query string) (driver.Stmt, error)
		CloseFunc   func() error
		BeginFunc   func() (driver.Tx, error)
	}
	mockTx struct {
		closed       bool
		CommitFunc   func() error
		RollbackFunc func() error
	}
	mockStmt struct {
		query        string
		closed       bool
		CloseFunc    func() error
		NumInputFunc func() int
		ExecFunc     func(args []driver.Value) (driver.Result, error)
		QueryFunc    func(args []driver.Value) (driver.Rows, error)
	}
	mockResult struct {
		LastInsertIDFunc func() (int64, error)
		RowsAffectedFunc func() (int64, error)
	}
	mockRows struct {
		query       string
		closed      bool
		ColumnsFunc func() []string
		CloseFunc   func() error
		NextFunc    func(dest []driver.Value) error
	}
	mockDatabase struct {
		PingFunc     func() error
		QueryFunc    func(query string, args ...interface{}) (*sql.Rows, error)
		QueryRowFunc func(query string, args ...interface{}) *sql.Row
		ExecFunc     func(query string, args ...interface{}) (sql.Result, error)
		BeginFunc    func() (*sql.Tx, error)
	}
)

func (m mockDriver) Open(name string) (driver.Conn, error) {
	return m.OpenFunc(name)
}

func (m mockConn) Prepare(query string) (driver.Stmt, error) {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(query)
	}
	return mockStmt{query: query}, nil
}
func (m mockConn) Close() error {
	if m.closed {
		return errors.New("mock connection already closed")
	}
	m.closed = false
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
func (m mockConn) Begin() (driver.Tx, error) {
	if m.closed {
		return nil, errors.New("mock connection already closed")
	}
	m.closed = true
	if m.BeginFunc != nil {
		return m.BeginFunc()
	}
	return mockTx{}, nil
}

func (m mockTx) Commit() error {
	if m.closed {
		return errors.New("mock transaction already closed")
	}
	m.closed = true
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return nil
}
func (m mockTx) Rollback() error {
	if m.closed {
		return errors.New("mock transaction already closed")
	}
	m.closed = true
	if m.RollbackFunc != nil {
		return m.RollbackFunc()
	}
	return nil
}

func (m mockStmt) Close() error {
	if m.closed {
		return errors.New("mock statement already closed")
	}
	m.closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
func (m mockStmt) NumInput() int {
	if m.NumInputFunc != nil {
		return m.NumInputFunc()
	}
	return 0
}
func (m mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	if m.closed {
		return nil, errors.New("mock statement already closed")
	}
	if m.ExecFunc != nil {
		return m.ExecFunc(args)
	}
	return mockResult{}, nil
}
func (m mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	if m.closed {
		return nil, errors.New("mock statement already closed")
	}
	if m.QueryFunc != nil {
		return m.QueryFunc(args)
	}
	return mockRows{query: m.query}, nil
}

func (m mockResult) LastInsertId() (int64, error) {
	if m.LastInsertIDFunc != nil {
		return m.LastInsertIDFunc()
	}
	return 0, nil
}
func (m mockResult) RowsAffected() (int64, error) {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0, nil
}

func (m mockRows) Columns() []string {
	if m.ColumnsFunc != nil {
		return m.ColumnsFunc()
	}
	return nil
}
func (m mockRows) Close() error {
	if m.closed {
		return errors.New("mock rows already closed")
	}
	m.closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
func (m mockRows) Next(dest []driver.Value) error {
	if m.closed {
		return errors.New("mock rows already closed")
	}
	m.closed = true
	if m.NextFunc != nil {
		return m.NextFunc(dest)
	}
	return nil
}

func (m mockDatabase) Ping() error {
	return m.PingFunc()
}
func (m mockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.QueryFunc(query, args)
}
func (m mockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.QueryRowFunc(query, args)
}
func (m mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args)
}
func (m mockDatabase) Begin() (*sql.Tx, error) {
	return m.BeginFunc()
}

func initializeWithTestDb(t *testing.T) {
	driverName := "mockDriverName"
	dataSourceName := "mockDataSourceName"
	mockConn := mockConn{}
	mockDriver := mockDriver{
		OpenFunc: func(name string) (driver.Conn, error) {
			if name != dataSourceName {
				return nil, fmt.Errorf("invalid driver name: %v", name)
			}
			return mockConn, nil
		},
	}
	sql.Register(driverName, mockDriver)
	mockDb, err := sql.Open(driverName, dataSourceName)
	switch {
	case err != nil:
		t.Fatal(err)
	case mockDb == nil:
		t.Fatal("no sql database")
	}
}

func TestInit_notImported(t *testing.T) {
	dataSourceName := "mockDataSourceName"
	err := Init(dataSourceName)
	if err == nil {
		t.Error("expected error")
	}
}

func TestInit_ok(t *testing.T) {
	mockDriver := mockDriver{}
	driverName := "postgres"
	dataSourceName := "mockDataSourceName"
	sql.Register(driverName, mockDriver)
	err := Init(dataSourceName)
	if err != nil {
		t.Error("unexpected error:", err)
	}
}

func TestExpectSingleRowAffected(t *testing.T) {
	expectSingleRowAffectedTests := []struct {
		rowsAffectedErr error
		rows            int64
		wantErr         bool
	}{
		{
			rowsAffectedErr: errors.New("could not get rows affected"),
			rows:            1, // red herring :)
			wantErr:         true,
		},
		{
			rows: 1, // desired
		},
		{
			rows:    0,
			wantErr: true,
		},
		{
			rows:    3,
			wantErr: true,
		},
	}
	for i, test := range expectSingleRowAffectedTests {
		r := mockResult{
			RowsAffectedFunc: func() (int64, error) {
				return test.rows, test.rowsAffectedErr
			},
		}
		err := expectSingleRowAffected(r)
		switch {
		case test.wantErr && err == nil:
			t.Errorf("Test %v: expected error", i)
		case !test.wantErr && err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		}
	}
}
