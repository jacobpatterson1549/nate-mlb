package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type (
	// mockDatabase implements the database interface
	mockDatabase struct {
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
		ScanFunc  func(dest ...interface{}) error
	}
	// mockTransaction implements the transaction interface
	mockTransaction struct {
		ExecFunc     func(query string, args ...interface{}) (sql.Result, error)
		CommitFunc   func() error
		RollbackFunc func() error
	}
	// mockResult implements the sql.Result interface
	mockResult struct {
		LastInsertIDFunc func() (int64, error)
		RowsAffectedFunc func() (int64, error)
	}
	// mockDriver implements the sql/driver/Driver interface
	mockDriver struct {
		OpenFunc func(name string) (driver.Conn, error)
	}
	// mockDriverConn implements the sql/driver/Conn interface
	mockDriverConn struct {
		PrepareFunc func(query string) (driver.Stmt, error)
		CloseFunc   func() error
		BeginFunc   func() (driver.Tx, error)
	}
	// mockDriverTx implements the sql/driver/Tx interface
	mockDriverTx struct {
		CommitFunc   func() error
		RollbackFunc func() error
	}
	// mockDriverStmt implements the sql/driver/Stmt interface
	mockDriverStmt struct {
		CloseFunc    func() error
		NumInputFunc func() int
		ExecFunc     func(args []driver.Value) (driver.Result, error)
		QueryFunc    func(args []driver.Value) (driver.Rows, error)
	}
	// mockDriverRows implements the sql/driver/Rows interface
	mockDriverRows struct {
		ColumnsFunc func() []string
		CloseFunc   func() error
		NextFunc    func(dest []driver.Value) error
	}
)

func (m mockDatabase) Query(query string, args ...interface{}) (rows, error) {
	return m.QueryFunc(query, args...)
}
func (m mockDatabase) QueryRow(query string, args ...interface{}) row {
	return m.QueryRowFunc(query, args...)
}
func (m mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args...)
}
func (m mockDatabase) Begin() (transaction, error) {
	return m.BeginFunc()
}
func (m mockRow) Scan(dest ...interface{}) error {
	return m.ScanFunc(dest...)
}
func (m mockRows) Close() error {
	return m.CloseFunc()
}
func (m mockRows) Next() bool {
	return m.NextFunc()
}
func (m mockRows) Scan(dest ...interface{}) error {
	return m.ScanFunc(dest...)
}
func (m mockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.ExecFunc(query, args...)
}
func (m mockTransaction) Commit() error {
	return m.CommitFunc()
}
func (m mockTransaction) Rollback() error {
	return m.RollbackFunc()
}
func (m mockResult) LastInsertId() (int64, error) {
	return m.LastInsertIDFunc()
}
func (m mockResult) RowsAffected() (int64, error) {
	return m.RowsAffectedFunc()
}
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
func (m mockDriverRows) Columns() []string {
	return m.ColumnsFunc()
}
func (m mockDriverRows) Close() error {
	return m.CloseFunc()
}
func (m mockDriverRows) Next(dest []driver.Value) error {
	return m.NextFunc(dest)
}

func mockScan(dest, src interface{}) error {
	switch s := src.(type) {
	case bool:
		switch d := dest.(type) {
		case *bool:
			*d = s
			return nil
		}
	case int:
		switch d := dest.(type) {
		case *int:
			*d = s
			return nil
		case *SportType:
			*d = SportType(s)
			return nil
		case *PlayerType:
			*d = PlayerType(s)
			return nil
		case *ID:
			*d = ID("s")
			return nil
		}
	case string:
		switch d := dest.(type) {
		case *string:
			*d = s
			return nil
		case *sql.NullString:
			d.Valid = true
			d.String = s
			return nil
		}
	case ID:
		switch d := dest.(type) {
		case *ID:
			*d = s
			return nil
		}
	case PlayerType:
		switch d := dest.(type) {
		case *PlayerType:
			*d = s
			return nil
		}
	case SourceID:
		switch d := dest.(type) {
		case *SourceID:
			*d = s
			return nil
		}
	case *time.Time:
		switch d := dest.(type) {
		case **time.Time:
			*d = s
			return nil
		}
	case *[]byte:
		switch d := dest.(type) {
		case **[]byte:
			*d = s
			return nil
		}
	case *sql.NullString:
		switch d := dest.(type) {
		case *sql.NullString:
			if s != nil {
				*d = *s
			}
			return nil
		}
	}
	return fmt.Errorf("expected %T for destination of scan, but was %T", dest, src)
}

func newMockRows(src []interface{}) rows {
	closed := false
	rowI := -1
	return mockRows{
		CloseFunc: func() error {
			if closed {
				return fmt.Errorf("already closed")
			}
			closed = true
			return nil
		},
		NextFunc: func() bool {
			rowI++
			return rowI < len(src)
		},
		ScanFunc: func(dest ...interface{}) error {
			switch {
			case closed:
				return fmt.Errorf("already closed")
			case rowI < 0:
				return fmt.Errorf("next not called")
			case rowI >= len(src):
				return fmt.Errorf("no more rows")
			}
			return mockRowScanFunc(src[rowI], dest...)
		},
	}
}

func mockRowScanFunc(src interface{}, dest ...interface{}) error {
	s := reflect.ValueOf(&src).Elem().Elem()
	dI := len(dest)
	sI := s.NumField()
	if dI != sI {
		return fmt.Errorf("dest has %v fields, yet src has %v", dI, sI)
	}
	for i := 0; i < dI; i++ {
		f := s.Field(i)
		sI := f.Interface()
		if err := mockScan(dest[i], sI); err != nil {
			return err
		}
	}
	return nil
}

func newMockBeginFunc(beginErr error, commitValidator func(queries []writeSQLFunction)) func() (transaction, error) {
	return func() (transaction, error) {
		var queries []writeSQLFunction
		r := mockResult{
			RowsAffectedFunc: func() (int64, error) {
				return 1, nil
			},
		}
		return mockTransaction{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				queries = append(queries, writeSQLFunction{name: query, args: args})
				return r, nil
			},
			CommitFunc: func() error {
				commitValidator(queries)
				return nil
			},
		}, beginErr
	}
}

func init() {
	sql.Register("TestNewSqlDatabaseDriver", mockDriver{})
}
func TestNewSqlDatabase(t *testing.T) {
	db, err := newSQLDatabase("TestNewSqlDatabaseDriver", "mockDataSourceName")
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if db == nil {
		t.Error("expected database to not be nil after Init() called, but was")
	}
}

func TestNewSqlDatabase_missingDriver(t *testing.T) {
	db, err := newSQLDatabase("missing driver", "mockDataSourceName")
	if err == nil {
		t.Error("expected error because driver is missing")
	}
	if db != nil {
		t.Errorf("did not expect database, but got %v", db)
	}
}

var testSQLDatabaseMethodsDriver *mockDriver

func init() {
	testSQLDatabaseMethodsDriver = new(mockDriver)
	sql.Register("TestSQLDatabaseMethods", testSQLDatabaseMethodsDriver)
}
func TestSQLDatabaseMethods(t *testing.T) {
	var pingCalled, queryCalled, queryRowCalled, execCalled, beginTransactionCalled int
	mockDriverConn := mockDriverConn{
		PrepareFunc: func(query string) (driver.Stmt, error) {
			return mockDriverStmt{
				CloseFunc: func() error {
					return nil
				},
				NumInputFunc: func() int {
					return 0
				},
				ExecFunc: func(args []driver.Value) (driver.Result, error) {
					execCalled++
					return mockResult{}, nil
				},
				QueryFunc: func(args []driver.Value) (driver.Rows, error) {
					switch query {
					case "query_sql":
						queryCalled++
					case "query_row_sql":
						queryRowCalled++
					}
					return mockDriverRows{}, nil
				},
			}, nil
		},
		BeginFunc: func() (driver.Tx, error) {
			beginTransactionCalled++
			return mockDriverTx{}, nil
		},
	}
	testSQLDatabaseMethodsDriver.OpenFunc = func(name string) (driver.Conn, error) {
		pingCalled++
		return mockDriverConn, nil
	}
	d, err := newSQLDatabase("TestSQLDatabaseMethods", "mockDataSourceName")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if d == nil {
		t.Fatal("expected database to not be nil after Init() called, but was")
	}
	methodTests := []struct {
		methodName string
		count      *int
		method     func()
	}{
		{
			methodName: "queryCalled",
			count:      &queryCalled,
			method:     func() { d.db.Query("query_sql") },
		},
		{
			methodName: "queryRowCalled",
			count:      &queryRowCalled,
			method:     func() { d.db.QueryRow("query_row_sql") },
		},
		{
			methodName: "execCalled",
			count:      &execCalled,
			method:     func() { d.db.Exec("exec_sql") },
		},
		{
			methodName: "beginTransactionCalled",
			count:      &beginTransactionCalled,
			method:     func() { d.db.Begin() },
		},
	}
	for _, call := range methodTests {
		if *call.count != 0 {
			t.Errorf("%v already called", call.methodName)
			continue
		}
		call.method()
		if *call.count != 1 {
			t.Errorf("%v not called", call.methodName)
		}
	}
	for i, call := range methodTests {
		if i != 0 && *call.count != 1 { // ignore extra Pings
			t.Errorf("%v called more than once: %v", call.methodName, *call.count)
		}
	}
}
