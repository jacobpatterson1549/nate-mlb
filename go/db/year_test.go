package db

import (
	"errors"
	"fmt"
	"testing"
)

func TestGetYears(t *testing.T) {
	getYearsTests := []struct {
		requestSportType SportType
		rowsSportType    SportType
		queryErr         error
		rows             []interface{}
		wantSlice        []Year
		wantErr          bool
	}{
		{},
		{
			queryErr: fmt.Errorf("query error"),
			wantErr:  true,
		},
		{ // incorrect sportType
			requestSportType: 1,
			rowsSportType:    2,
			rows: []interface{}{
				Year{
					Value:  2019,
					Active: true,
				},
			},
		},
		{ // happy path
			requestSportType: 3,
			rowsSportType:    3,
			rows: []interface{}{
				Year{
					Value:  2017,
					Active: false,
				},
				Year{
					Value:  2019,
					Active: true,
				},
				Year{
					Value:  2018,
					Active: false,
				},
			},
			wantSlice: []Year{
				Year{
					Value:  2017,
					Active: false,
				},
				Year{
					Value:  2019,
					Active: true,
				},
				Year{
					Value:  2018,
					Active: false,
				},
			},
		},
		{ // multiple active
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				Year{
					Value:  2020,
					Active: true,
				},
				Year{
					Value:  2019,
					Active: true,
				},
			},
			wantErr: true,
		},
		{ // scan error
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				struct {
					Value  string
					Active int
				}{
					Value:  "2020",
					Active: 1,
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getYearsTests {
		db = mockDatabase{
			QueryFunc: func(query string, args ...interface{}) (rows, error) {
				if test.queryErr != nil {
					return nil, test.queryErr
				}
				if test.requestSportType != test.rowsSportType {
					return newMockRows([]interface{}{}), nil
				}
				return newMockRows(test.rows), nil
			},
		}
		gotSlice, gotErr := GetYears(test.requestSportType)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case test.requestSportType != test.rowsSportType:
			if len(gotSlice) != 0 {
				t.Errorf("Test %v: expected no rows for incorrect sportType, but got %v", i, gotSlice)
			}
		default:
			if len(test.rows) != len(gotSlice) {
				t.Errorf("Test %v: incorrect output rows", i)
			}
			for j, got := range gotSlice {
				want := test.wantSlice[j]
				if want != got {
					t.Errorf("Test %v, %T %v not equal: want %v, got %v", i, j, want, want, got)
				}
			}
		}
	}
}

func TestSaveYears(t *testing.T) {
	saveYearsTests := []struct {
		st                      SportType
		futureYears             []Year
		previousYears           []Year
		getYearsErr             error
		executeInTransactionErr error
		wantErr                 bool
		wantQueryYears          []int
	}{
		{},
		{ // happy path
			st: 8,
			futureYears: []Year{
				{
					Value:  2019,
					Active: true,
				},
				{
					Value:  2020,
					Active: false,
				},
				{
					Value:  2018,
					Active: false,
				},
			},
			previousYears: []Year{
				{
					Value:  2017,
					Active: false,
				},
				{
					Value:  2018,
					Active: true,
				},
			},
			wantQueryYears: []int{2017, 2019, 2020, 2019},
		},
		{
			getYearsErr: errors.New("getYears error"),
		},
		{
			executeInTransactionErr: errors.New("executeInTransaction error"),
		},
		{ // multiple active years
			futureYears: []Year{
				{
					Value:  2019,
					Active: true,
				},
				{
					Value:  2020,
					Active: true,
				},
			},
			wantErr: true,
		},
	}
	for i, test := range saveYearsTests {
		getYearsFunc := func(st SportType) ([]Year, error) {
			if test.st != st {
				t.Errorf("Test %v: wanted to get years for SportType %v, but got %v", i, test.st, st)
			}
			return test.previousYears, test.getYearsErr
		}
		executeInTransactionFunc := func(queries []writeSQLFunction) error {
			// first query is to clear active year, then delete years, insert years, and set active
			if len(test.wantQueryYears)+1 != len(queries) {
				t.Errorf("Test %v: wanted %v queries, got %v", i, len(test.wantQueryYears)+1, len(queries))
			}
			for i, wantQueryYear := range test.wantQueryYears {
				switch v := queries[i+1].args[1].(type) { // args should be []{st, Year.Value}
				case int:
					gotQueryYear := v
					if wantQueryYear != gotQueryYear {
						t.Errorf("Test %v: wanted param 2 of query %v to be %v, got %v", i, i+1, wantQueryYear, gotQueryYear)
					}
				default:
					t.Errorf("Test %v: wanted param 2 of query %v to be an int (year), got %v", i, i+1, v)
				}
			}
			return test.executeInTransactionErr
		}
		wantErr := test.wantErr || test.getYearsErr != nil || test.executeInTransactionErr != nil
		gotErr := saveYears(test.st, test.futureYears, getYearsFunc, executeInTransactionFunc)
		switch {
		case wantErr && gotErr == nil:
			t.Errorf("Test %v: wanted error", i)
		case !wantErr && gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case test.getYearsErr != nil && !errors.Is(gotErr, test.getYearsErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.getYearsErr, gotErr)
		case test.executeInTransactionErr != nil && !errors.Is(gotErr, test.executeInTransactionErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.executeInTransactionErr, gotErr)
		}
	}
}
