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
	mockDriverConn struct {
		closed      bool
		PrepareFunc func(query string) (driver.Stmt, error)
		CloseFunc   func() error
		BeginFunc   func() (driver.Tx, error)
	}
	mockDriverTx struct {
		CommitFunc   func() error
		RollbackFunc func() error
	}
	mockDriverStmt struct {
		CloseFunc    func() error
		NumInputFunc func() int
		ExecFunc     func(args []driver.Value) (driver.Result, error)
		QueryFunc    func(args []driver.Value) (driver.Rows, error)
	}
	mockDriverResult struct {
		LastInsertIDFunc func() (int64, error)
		RowsAffectedFunc func() (int64, error)
	}
	mockDriverRows struct {
		ColumnsFunc func() []string
		CloseFunc   func() error
		NextFunc    func(dest []driver.Value) error
	}
)

func (m mockDriver) Open(name string) (driver.Conn, error) {
	return m.OpenFunc(name)
}

func (m mockDriverConn) Prepare(query string) (driver.Stmt, error) {
	return m.PrepareFunc(query)
}
func (m mockDriverConn) Close() error {
	return m.CloseFunc()
}
func (m mockDriverConn) Begin() (driver.Tx, error) {
	return m.BeginFunc()
}

func (m mockDriverTx) Commit() error {
	return m.CommitFunc()
}
func (m mockDriverTx) Rollback() error {
	return m.RollbackFunc()
}

func (m mockDriverStmt) Close() error {
	return m.CloseFunc()
}
func (m mockDriverStmt) NumInput() int {
	return m.NumInputFunc()
}
func (m mockDriverStmt) Exec(args []driver.Value) (driver.Result, error) {
	return m.ExecFunc(args)
}
func (m mockDriverStmt) Query(args []driver.Value) (driver.Rows, error) {
	return m.QueryFunc(args)
}

func (m mockDriverResult) LastInsertId() (int64, error) {
	return m.LastInsertIDFunc()
}
func (m mockDriverResult) RowsAffected() (int64, error) {
	return m.RowsAffectedFunc()
}

func (m mockDriverRows) Columns() []string {
	return nil
}
func (m mockDriverRows) Close() error {
	return nil
}
func (m mockDriverRows) Next(dest []driver.Value) error {
	return nil
}

func init() {
	driverName, dataSourceName := "postgres", "mockDataSourceName"
	mockConn := mockDriverConn{}
	mockDriver := mockDriver{
		OpenFunc: func(name string) (driver.Conn, error) {
			if name != dataSourceName {
				return nil, fmt.Errorf("invalid dataSourceName: %v", name)
			}
			return mockConn, nil
		},
	}
	sql.Register(driverName, mockDriver)
}

func TestInit_ok(t *testing.T) {
	dataSourceName := "mockDataSourceName"
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
