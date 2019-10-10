package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestGetStat(t *testing.T) {
	type sportTypeQueryRow struct {
		Year         int
		EtlTimestamp *time.Time
		EtlJSON      *[]byte
	}
	testTime := time.Date(2019, time.October, 10, 10, 17, 33, 0, time.UTC)
	getStatTests := []struct {
		requestSportType SportType
		rowSportType     SportType
		queryRowErr      error
		row              interface{}
		wantStat         *Stat
		wantErr          bool
	}{
		{},
		{
			queryRowErr: errors.New("queryRow error"),
			wantErr:     true,
		},
		{ // incorrect sportType
			requestSportType: 1,
			rowSportType:     2,
		},
		{ // happy path
			requestSportType: 8,
			rowSportType:     8,
			row: sportTypeQueryRow{
				Year:         2019,
				EtlTimestamp: &testTime,
			},
			wantStat: &Stat{
				SportType:    8,
				Year:         2019,
				EtlTimestamp: &testTime,
				EtlJSON:      nil,
			},
		},
	}
	for i, test := range getStatTests {
		db = mockDatabase{
			QueryRowFunc: func(query string, args ...interface{}) row {
				return mockRow{
					ScanFunc: func(dest ...interface{}) error {
						switch {
						case test.queryRowErr != nil:
							return test.queryRowErr
						case test.requestSportType != test.rowSportType,
							test.row == nil:
							return sql.ErrNoRows
						default:
							return mockRowScanFunc(test.row, dest...)
						}
					},
				}
			},
		}
		gotStat, gotErr := GetStat(test.requestSportType)
		switch {
		case test.wantErr:
			switch {
			case gotErr == nil:
				t.Errorf("Test %v: expected error", i)
			case test.queryRowErr != nil && !errors.Is(gotErr, test.queryRowErr):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.queryRowErr, gotErr)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case test.requestSportType != test.rowSportType:
			if gotStat != nil {
				t.Errorf("Test %v: expected nil Stat for incorrect sportType, but got %v", i, gotStat)
			}
		case test.wantStat == nil:
			if gotStat != nil {
				t.Errorf("Test %v: wanted nil stat, but to %v", i, gotStat)
			}
		default:
			if *test.wantStat != *gotStat {
				t.Errorf("Test %v, not equal: want %v, got %v", i, test.wantStat, gotStat)
			}
			if test.requestSportType != gotStat.SportType {
				t.Errorf("Test %v: wanted SportType to be %v (requested SportType); got %v", i, test.requestSportType, gotStat.SportType)
			}
		}
	}
}

func TestSetStat(t *testing.T) {
	setStatTests := []struct {
		stat         Stat
		rowsAffected int64
		execError    error
		wantErr      bool
	}{
		{
			rowsAffected: 1,
		},
		{
			execError: errors.New("queryRow error"),
			wantErr:   true,
		},
		{
			rowsAffected: 4,
			wantErr:      true,
		},
	}
	for i, test := range setStatTests {
		db = mockDatabase{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				if test.execError != nil {
					return nil, test.execError
				}
				return mockResult{
					RowsAffectedFunc: func() (int64, error) {
						return test.rowsAffected, nil
					},
				}, nil
			},
		}
		gotErr := SetStat(test.stat)
		switch {
		case test.wantErr:
			switch {
			case gotErr == nil:
				t.Errorf("Test %v: expected error", i)
			case test.execError != nil && !errors.Is(gotErr, test.execError):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.execError, gotErr)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestClearStat(t *testing.T) {
	clearStatTests := []struct {
		st        SportType
		execError error
		wantErr   bool
	}{
		{},
		{
			execError: errors.New("queryRow error"),
			wantErr:   true,
		},
	}
	for i, test := range clearStatTests {
		db = mockDatabase{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				return nil, test.execError
			},
		}
		gotErr := ClearStat(test.st)
		switch {
		case test.wantErr:
			switch {
			case gotErr == nil:
				t.Errorf("Test %v: expected error", i)
			case test.execError != nil && !errors.Is(gotErr, test.execError):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.execError, gotErr)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}
