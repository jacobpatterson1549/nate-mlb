package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"
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
		RowsAffectedFunc func() (int64, error)
	}
)

func (m mockDatabase) Ping() error {
	return m.PingFunc()
}
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
	panic("should not be called") // required by sql.Result interface
}
func (m mockResult) RowsAffected() (int64, error) {
	return m.RowsAffectedFunc()
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
			*d = ID(s)
			return nil
		}
	case string:
		switch d := dest.(type) {
		case *string:
			*d = s
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

func newMockBeginFunc(commitValidator func(queries []writeSQLFunction) error) func() (transaction, error) {
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
				return commitValidator(queries)
			},
		}, nil
	}
}

func TestNewSqlDatabase(t *testing.T) {
	dataSourceName := "mockDataSourceName"
	db, err := newSQLDatabase(dataSourceName)
	switch {
	case err != nil:
		t.Error("unexpected error:", err)
	case db == nil:
		t.Error("expected database to not be nil after Init() called, but was")
	}
}
