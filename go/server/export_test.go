package server

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestExportToCsv(t *testing.T) {
	es := EtlStats{
		sportTypeName: "rugby",
		year:          2008,
	}
	var w bytes.Buffer
	err := exportToCsv(es, &w)
	want := "nate-mlb,2008 rugby scores\n\ntype,friend,value,player,score\n"
	got := w.String()
	switch {
	case err != nil:
		t.Errorf("did not expect error: %v", err)
	case want != got:
		t.Errorf("different csv:\nwanted: %v\ngot:    %v", want, got)
	}
}

type errWriter struct {
	err error
}

func (w errWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestExportToCsv_writeError(t *testing.T) {
	err := errors.New("write failed")
	var es EtlStats
	w := errWriter{err: err}
	got := exportToCsv(es, w)
	if !errors.Is(got, err) {
		t.Errorf("did not get expected export error: wanted: %v, got: %v", err, got)
	}
}

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
