package server

import (
	"fmt"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/internal/request"
)

func TestCreateCsvRecords(t *testing.T) {
	es := EtlStats{
		ScoreCategories: []request.ScoreCategory{
			{
				Name: "teams",
				FriendScores: []request.FriendScore{
					{
						Name: "Arnold",
						PlayerScores: []request.PlayerScore{
							{Name: "San Francisco 49ers", Score: 4},
							{Name: "Arizona Cardinals", Score: 3},
						},
						Score: 7,
					},
					{
						Name: "Bert",
						PlayerScores: []request.PlayerScore{
							{Name: "Green Bay Packers", Score: 6},
							{Name: "Cleveland Browns", Score: 7},
						},
						Score: 13,
					},
				},
			},
			{
				Name: "qb",
				FriendScores: []request.FriendScore{
					{
						Name: "Charlie",
						PlayerScores: []request.PlayerScore{
							{Name: "Tom Brady", Score: 29},
						},
						Score: 29,
					},
				},
			},
		},
		sportTypeName: "american football",
		year:          2018,
	}
	got := createCsvRecords(es)

	want := [][]string{
		{"nate-mlb", "2018 american football scores"},
		nil,
		{"type", "friend", "value", "player", "score"},
		nil,
		{"teams", "Arnold", "7", "San Francisco 49ers", "4"},
		{"", "", "", "Arizona Cardinals", "3"},
		nil,
		{"", "Bert", "13", "Green Bay Packers", "6"},
		{"", "", "", "Cleveland Browns", "7"},
		nil,
		nil,
		{"qb", "Charlie", "29", "Tom Brady", "29"},
	}

	if err := equal2DArrays(want, got); err != nil {
		t.Errorf("different CSV: %v", err)
	}
}

func equal2DArrays(want, got [][]string) error {
	if len(want) != len(got) {
		return fmt.Errorf("row counts: wanted %d, got %d", len(want), len(got))
	}
	for i := range want {
		wantRow, gotRow := want[i], got[i]
		if len(wantRow) != len(gotRow) {
			return fmt.Errorf("column counts at row %d: wanted %d, got %d", i, len(wantRow), len(gotRow))
		}
		for j := range wantRow {
			wantValue, gotValue := wantRow[j], gotRow[j]
			if wantValue != gotValue {
				return fmt.Errorf("values at row %d, column %d: %s, %s", i, j, wantValue, gotValue)
			}
		}
	}
	return nil
}
