package db

import (
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
