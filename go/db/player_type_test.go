package db

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPlayerTypeName(t *testing.T) {
	want := "team"
	pt := PlayerType(1)
	playerTypes[pt] = playerType{name: want}
	got := pt.Name()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestPlayerTypeDescription(t *testing.T) {
	want := "people"
	pt := PlayerType(2)
	playerTypes[pt] = playerType{description: want}
	got := pt.Description()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestPlayerTypeScoreType(t *testing.T) {
	want := "wins"
	pt := PlayerType(4)
	playerTypes[pt] = playerType{scoreType: want}
	got := pt.ScoreType()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestPlayerTypeDisplayOrder(t *testing.T) {
	want := 7
	pt := PlayerType(4)
	playerTypes[pt] = playerType{displayOrder: want}
	got := pt.DisplayOrder()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestPlayerTypes(t *testing.T) {
	pt1 := PlayerType(1)
	pt2 := PlayerType(2)
	pt3 := PlayerType(3)
	st1 := SportType(1)
	st2 := SportType(2)
	playerTypes = map[PlayerType]playerType{
		pt1: {sportType: st1, displayOrder: 2},
		pt2: {sportType: st2, displayOrder: 0},
		pt3: {sportType: st1, displayOrder: 1},
	}
	want := []PlayerType{
		pt3,
		pt1,
	}
	got := PlayerTypes(st1)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestLoadPlayerTypes(t *testing.T) {
	type playerTypeQueryRow struct {
		ID          int
		SportType   int
		Name        string
		Description string
		ScoreType   string
	}
	loadPlayerTypesTests := []struct {
		queryErr error
		rows     []interface{}
		wantErr  bool
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
				playerTypeQueryRow{5, 2, "mockNflQBName", "mockNflQBDescription", "mockNflQBScoreType"},
				playerTypeQueryRow{6, 2, "mockNflMiscName", "mockNflMiscDescription", "mockNflMiscType"},
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
	for i, test := range loadPlayerTypesTests {
		db = mockDatabase{
			QueryFunc: func(query string, args ...interface{}) (rows, error) {
				if test.queryErr != nil {
					return nil, test.queryErr
				}
				return newMockRows(test.rows), nil
			},
		}
		gotErr := LoadPlayerTypes()
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}
