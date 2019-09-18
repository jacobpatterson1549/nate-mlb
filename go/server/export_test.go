package server

import (
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestCreateCsvRecords(t *testing.T) {
	es := EtlStats{
		scoreCategories: []request.ScoreCategory{
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

	if !reflect.DeepEqual(want, got) {
		t.Errorf("different CSV:\nwanted: %v\ngot:    %v", want, got)
	}
}
