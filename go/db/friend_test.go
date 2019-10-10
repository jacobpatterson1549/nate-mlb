package db

import (
	"fmt"
	"testing"
)

func TestGetFriends(t *testing.T) {
	getFriendsTests := []struct {
		requestSportType SportType
		rowsSportType    SportType
		queryErr         error
		rows             []interface{}
		wantSlice        []Friend
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
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "arnold",
				},
			},
		},
		{ // happy path
			requestSportType: 3,
			rowsSportType:    3,
			rows: []interface{}{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "alfred",
				},
				Friend{
					ID:           6,
					DisplayOrder: 3,
					Name:         "aaron",
				},
				Friend{
					ID:           4,
					DisplayOrder: 2,
					Name:         "earl",
				},
			},
			wantSlice: []Friend{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "alfred",
				},
				Friend{
					ID:           6,
					DisplayOrder: 3,
					Name:         "aaron",
				},
				Friend{
					ID:           4,
					DisplayOrder: 2,
					Name:         "earl",
				},
			},
		},
		{ // scan error
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				struct {
					ID           string
					DisplayOrder int
					Name         string
				}{
					ID:           "1",
					DisplayOrder: 1,
					Name:         "arnold",
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getFriendsTests {
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
		gotSlice, gotErr := GetFriends(test.requestSportType)
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
