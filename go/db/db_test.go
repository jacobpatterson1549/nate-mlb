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

func TestPing(t *testing.T) {
	pingTests := []struct {
		pingErr error
	}{
		{},
		{
			pingErr: fmt.Errorf("Ping error"),
		},
	}
	for i, test := range pingTests {
		db = mockDatabase{
			PingFunc: func() error {
				return test.pingErr
			},
		}
		wantErr := test.pingErr
		gotErr := Ping()
		if wantErr != gotErr {
			t.Errorf("Test %v: wanted %v, got %v", i, wantErr, gotErr)
		}
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

func TestExpectRowFound(t *testing.T) {
	expectRowFoundTests := []struct {
		found   bool
		scanErr error
		wantErr bool
	}{
		{
			found: true,
		},
		{
			scanErr: fmt.Errorf("scanError"),
		},
		{
			found:   false,
			wantErr: true,
		},
	}
	for i, test := range expectRowFoundTests {
		r := mockRow{
			ScanFunc: func(dest ...interface{}) error {
				if test.scanErr != nil {
					return test.scanErr
				}
				switch d := dest[0].(type) {
				case *bool:
					*d = test.found
					return nil
				default:
					return fmt.Errorf("Expected *bool for destination of scan, but was %T", dest[0])
				}
			},
		}
		gotErr := expectRowFound(r)
		switch {
		case gotErr == nil && (test.scanErr != nil || test.wantErr):
			t.Errorf("Test %v: expected error", i)
		case gotErr != nil && test.found:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}
