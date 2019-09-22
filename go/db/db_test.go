package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
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
		closed      bool
		ColumnsFunc func() []string
		CloseFunc   func() error
		NextFunc    func(dest []driver.Value) error
	}

	mockFileInfo struct {
		NameFunc    func() string
		SizeFunc    func() int64
		ModeFunc    func() os.FileMode
		ModTimeFunc func() time.Time
		IsDirFunc   func() bool
		SysFunc     func() interface{}
	}
)

func (m mockDriver) Open(name string) (driver.Conn, error) {
	return m.OpenFunc(name)
}

func (m mockConn) Prepare(query string) (driver.Stmt, error) {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(query)
	}
	return mockStmt{}, nil
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
	return mockRows{}, nil
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

func (m mockFileInfo) Name() string {
	return m.NameFunc()
}
func (m mockFileInfo) Size() int64 {
	return m.SizeFunc()
}
func (m mockFileInfo) Mode() os.FileMode {
	return m.ModeFunc()
}
func (m mockFileInfo) ModTime() time.Time {
	return m.ModTimeFunc()
}
func (m mockFileInfo) IsDir() bool {
	return m.IsDirFunc()
}
func (m mockFileInfo) Sys() interface{} {
	return m.SysFunc()
}

func TestInit(t *testing.T) {
	driverName := "mockDriverName"
	dataSourceName := "mockDataSourceName"
	if db != nil {
		t.Fatal("database already initialized")
	}
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

	getSetupFileContents = func(filename string) ([]byte, error) {
		return []byte(filename + ";a;b;c"), nil
	}
	getSetupFunctionDirContents = func(dirname string) ([]os.FileInfo, error) {
		mockFileInfos := make([]os.FileInfo, 3)
		for i := range mockFileInfos {
			mockFileInfos[i] = mockFileInfo{
				NameFunc: func() string {
					return fmt.Sprintf("%s/function_%d", dirname, i)
				},
			}
		}
		return mockFileInfos, nil
	}
	err := Init(driverName, dataSourceName)
	if err != nil {
		t.Errorf("failed to init database: %v", err)
	}
}
