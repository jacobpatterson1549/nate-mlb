package db

import (
	"fmt"
	"testing"
)

func TestGetPlayers(t *testing.T) {
	getPlayersTests := []struct {
		requestSportType SportType
		rowsSportType    SportType
		queryErr         error
		rows             []interface{}
		wantSlice        []Player
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
				Player{
					ID:           1,
					PlayerType:   1,
					SourceID:     1,
					FriendID:     1,
					DisplayOrder: 1,
				},
			},
		},
		{ // happy path
			requestSportType: 3,
			rowsSportType:    3,
			rows: []interface{}{
				Player{
					ID:           1,
					PlayerType:   1,
					SourceID:     1,
					FriendID:     1,
					DisplayOrder: 1,
				},
				Player{
					ID:           17,
					PlayerType:   3,
					SourceID:     6,
					FriendID:     2,
					DisplayOrder: 3,
				},
				Player{
					ID:           34,
					PlayerType:   3,
					SourceID:     4000,
					FriendID:     2,
					DisplayOrder: 2,
				},
			},
			wantSlice: []Player{
				Player{
					ID:           1,
					PlayerType:   1,
					SourceID:     1,
					FriendID:     1,
					DisplayOrder: 1,
				},
				Player{
					ID:           17,
					PlayerType:   3,
					SourceID:     6,
					FriendID:     2,
					DisplayOrder: 3,
				},
				Player{
					ID:           34,
					PlayerType:   3,
					SourceID:     4000,
					FriendID:     2,
					DisplayOrder: 2,
				},
			},
		},
		{ // scan error
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				struct {
					ID           ID
					PlayerType   PlayerType
					SourceID     SourceID
					FriendID     SourceID // should be ID
					DisplayOrder int
				}{
					ID:           1,
					PlayerType:   1,
					SourceID:     1,
					FriendID:     1,
					DisplayOrder: 1,
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getPlayersTests {
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
		gotSlice, gotErr := GetPlayers(test.requestSportType)
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
