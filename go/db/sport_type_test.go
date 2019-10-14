package db

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetSportTypes(t *testing.T) {
	type sportTypeQueryRow struct {
		ID   int
		Name string
		URL  string
	}
	getSportTypesTests := []struct {
		queryErr       error
		rows           []interface{}
		wantErr        bool
		wantSportTypes map[SportType]SportTypeInfo
	}{
		{
			queryErr: fmt.Errorf("query error"),
			wantErr:  true,
		},
		{ // no sport types
			wantErr: true,
		},
		{ // scan error
			rows: []interface{}{
				struct {
					ID   string
					Name string
					URL  string
				}{
					ID:   "1", // not an int
					Name: "mockMlbName",
					URL:  "mockMlbUrl",
				},
			},
			wantErr: true,
		},
		{ // happy path
			rows: []interface{}{
				sportTypeQueryRow{
					ID:   1,
					Name: "mockMlbName",
					URL:  "mockMlbUrl",
				},
				sportTypeQueryRow{
					ID:   2,
					Name: "mockNflName",
					URL:  "mockNflUrl",
				},
			},
		},
		{ // no nfl sportType
			rows: []interface{}{
				sportTypeQueryRow{
					ID:   1,
					Name: "mockMlbName",
					URL:  "mockMlbUrl",
				},
				sportTypeQueryRow{
					ID:   3,
					Name: "mockRugbyname",
					URL:  "mockRugbyurl",
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getSportTypesTests {
		db = mockDatabase{
			QueryFunc: func(query string, args ...interface{}) (rows, error) {
				if test.queryErr != nil {
					return nil, test.queryErr
				}
				return newMockRows(test.rows), nil
			},
		}
		gotSportTypes, gotErr := GetSportTypes()
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
			if !reflect.DeepEqual(test.wantSportTypes, gotSportTypes) {
				t.Errorf("Test %v: wanted %v sportTypes, got %v", i, test.rows, gotSportTypes)
			}
		}
	}
}
