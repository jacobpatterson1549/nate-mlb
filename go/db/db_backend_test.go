package db

import (
	"database/sql"
	"fmt"
	"reflect"
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
		LastInsertIDFunc func() (int64, error)
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
	return m.LastInsertIDFunc()
}
func (m mockResult) RowsAffected() (int64, error) {
	return m.RowsAffectedFunc()
}

func mockScan(dest, src interface{}) error {
	switch src.(type) {
	case bool:
		switch d := dest.(type) {
		case *bool:
			*d = reflect.ValueOf(src).Bool()
			return nil
		}
	case int:
		v := reflect.ValueOf(src).Int()
		switch d := dest.(type) {
		case *int:
			*d = int(v)
			return nil
		case *SportType:
			*d = SportType(v)
			return nil
		case *PlayerType:
			*d = PlayerType(v)
			return nil
		case *ID:
			*d = ID(v)
			return nil
		}
	case string:
		switch d := dest.(type) {
		case *string:
			*d = reflect.ValueOf(src).String()
			return nil
		}
	case ID:
		switch d := dest.(type) {
		case *ID:
			*d = ID(reflect.ValueOf(src).Int())
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
			s := reflect.ValueOf(&src[rowI]).Elem().Elem()
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
		},
	}
}
