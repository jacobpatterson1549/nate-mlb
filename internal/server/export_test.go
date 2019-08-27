package server

import (
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"testing"
)

func TestCreateCsvRecords(t *testing.T) {
	es := EtlStats{
		ScoreCategories: []request.ScoreCategory{
			{
				Name: "teams",
				FriendScores: []request.FriendScore{
					{
						FriendName: "Arnie",
						PlayerScores: []request.PlayerScore{
							{PlayerName: "San Francisco 49ers", Score: 4},
							{PlayerName: "Arizona Cardinals", Score: 3},
						},
						Score: 7,
					},
					{
						FriendName: "Bert",
						PlayerScores: []request.PlayerScore{
							{PlayerName: "Green Bay Packers", Score: 6},
							{PlayerName: "Cleveland Browns", Score: 7},
						},
						Score: 13,
					},
				},
			},
			{
				Name: "qb",
				FriendScores: []request.FriendScore{
					{
						FriendName: "Charlie",
						PlayerScores: []request.PlayerScore{
							{PlayerName: "Tom Brady", Score: 29},
						},
						Score: 29,
					},
				},
			},
		},
		sportType: db.SportTypeNfl,
		year:      2018,
	}
	got := createCsvRecords(es)

	want := [][]string{
		{"nate-mlb", "2018 NFL scores"},
		nil,
		{"type", "friend", "value", "player", "score"},
		nil,
		{"teams", "Arnie", "7", "San Francisco 49ers", "4"},
		{"", "", "", "Arizona Cardinals", "3"},
		nil,
		{"", "Bert", "13", "Green Bay Packers", "6"},
		{"", "", "", "Cleveland Browns", "7"},
		nil,
		nil,
		{"qb", "Charlie", "29", "Tom Brady", "29"},
	}

	if len(want) != len(got) {
		t.Errorf("different csv row counts: wanted %d, got %d", len(want), len(got))
	} else {
	for i := range want {
		wantRow, gotRow := want[i], got[i]
		if len(wantRow) != len(gotRow) {
			t.Errorf("different csv column counts at row %d: wanted %d, got %d", i, len(wantRow), len(gotRow))
		} else {
		for j := range wantRow {
			wantValue, gotValue := wantRow[j], gotRow[j]
			if wantValue != gotValue {
				t.Errorf("different csv values at row %d, column %d: %s, %s", i, j, wantValue, gotValue)
			}
		}
	}
	}
}
}
