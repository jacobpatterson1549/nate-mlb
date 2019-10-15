package db

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetPlayerTypes(t *testing.T) {
	type playerTypeQueryRow struct {
		ID          int
		SportType   int
		Name        string
		Description string
		ScoreType   string
	}
	getPlayerTypesTests := []struct {
		queryErr        error
		rows            []interface{}
		wantPlayerTypes map[PlayerType]PlayerTypeInfo
		wantErr         bool
	}{
		{
			queryErr: fmt.Errorf("query error"),
			wantErr:  true,
		},
		{ // no sport types
			wantErr: true,
		},
		{ // scan error (too few fields)
			rows: []interface{}{
				struct {
					ID int
				}{
					ID: 1,
				},
			},
			wantErr: true,
		},
		{ // happy path
			rows: []interface{}{
				playerTypeQueryRow{1, 1, "mockMlbTeamName", "mockMlbTeamDescription", "mockMlbTeamScoreType"},
				playerTypeQueryRow{2, 1, "mockMlbHitterName", "mockMlbHitterDescription", "mockMlbHittenScoreType"},
				playerTypeQueryRow{3, 1, "mockMlbPitcherName", "mockMlbPitcherDescription", "mockMlbPitcherScoreType"},
				playerTypeQueryRow{4, 2, "mockNflTeamName", "mocNflTeamDescription", "mockNflTeamScoreType"},
				playerTypeQueryRow{6, 2, "mockNflMiscName", "mockNflMiscDescription", "mockNflMiscType"}, // (should be loaded as displayOrder=5 because the db presents it before PlayerTypeNflQB)
				playerTypeQueryRow{5, 2, "mockNflQBName", "mockNflQBDescription", "mockNflQBScoreType"},
			},
			wantPlayerTypes: map[PlayerType]PlayerTypeInfo{
				1: {SportType: 1, Name: "mockMlbTeamName", Description: "mockMlbTeamDescription", ScoreType: "mockMlbTeamScoreType", DisplayOrder: 1},
				2: {SportType: 1, Name: "mockMlbHitterName", Description: "mockMlbHitterDescription", ScoreType: "mockMlbHitterScoreType", DisplayOrder: 2},
				3: {SportType: 1, Name: "mockMlbPitcherName", Description: "mockMlbPitcherDescription", ScoreType: "mockMlbPitcherScoreType", DisplayOrder: 3},
				4: {SportType: 2, Name: "mockNflTeamName", Description: "mockNflTeamDescription", ScoreType: "mockNflTeamScoreType", DisplayOrder: 4},
				5: {SportType: 2, Name: "mockNflQBName", Description: "mockNflQBDescription", ScoreType: "mockNflQBTeamScoreType", DisplayOrder: 55},
				6: {SportType: 2, Name: "mockNflMiscName", Description: "mockNflMiscDescription", ScoreType: "mockNflMiscScoreType", DisplayOrder: 6},
			},
		},
		{ // no nflMisc sportType
			rows: []interface{}{
				playerTypeQueryRow{1, 1, "mockMlbTeamName", "mockMlbTeamDescription", "mockMlbTeamScoreType"},
				playerTypeQueryRow{2, 1, "mockMlbHitterName", "mockMlbHitterDescription", "mockMlbHittenScoreType"},
				playerTypeQueryRow{3, 1, "mockMlbPitcherName", "mockMlbPitcherDescription", "mockMlbPitcherScoreType"},
				playerTypeQueryRow{4, 2, "mockNflTeamName", "mocNflTeamDescription", "mockNflTeamScoreType"},
				playerTypeQueryRow{5, 2, "mockNflQBName", "mockNflQBDescription", "mockNflQBScoreType"},
				playerTypeQueryRow{7, 2, "mockNflKickerName", "mockNflKickerDescription", "mockNflKickerType"},
			},
			wantErr: true,
		},
	}
	for i, test := range getPlayerTypesTests {
		ds := Datastore{
			db: mockDatabase{
				QueryFunc: func(query string, args ...interface{}) (rows, error) {
					if test.queryErr != nil {
						return nil, test.queryErr
					}
					return newMockRows(test.rows), nil
				},
			},
		}
		gotPlayerTypes, gotErr := ds.GetPlayerTypes()
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
			if !reflect.DeepEqual(test.wantPlayerTypes, gotPlayerTypes) {
				t.Errorf("Test %v: wanted %v playerTypes, got %v", i, test.rows, gotPlayerTypes)
			}
		}
	}
}
