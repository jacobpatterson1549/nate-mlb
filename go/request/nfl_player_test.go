package request

import (
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestNflPlayerRequestScoreCategory(t *testing.T) {
	requestScoreCategoryTests := []struct {
		pt          db.PlayerType
		friends     []db.Friend
		players     []db.Player
		playersJSON string
		wantErr     bool
		want        ScoreCategory
	}{
		{
			wantErr: true, // no playersJSON
		},
		{
			pt:      db.PlayerTypeNflQB,
			friends: []db.Friend{{ID: 2, DisplayOrder: 1, Name: "Carl"}},
			players: []db.Player{
				{ID: 3, SourceID: 2532975, FriendID: 2, DisplayOrder: 1}, // Russell Wilson 6
			},
			playersJSON: `{"players":[
				{"id":"2532975","name":"Russell Wilson","position":"QB","stats":{"1":"2","6":"5"}}]}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeNflQB,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 2, Name: "Carl", Score: 5,
						PlayerScores: []PlayerScore{
							{ID: 3, Name: "Russell Wilson", Score: 5, DisplayOrder: 1, SourceID: 2532975},
						},
					},
				},
			},
		},
		{
			pt:      db.PlayerTypeNflMisc,
			friends: []db.Friend{{ID: 8, DisplayOrder: 1, Name: "Dave"}},
			players: []db.Player{
				{ID: 1, SourceID: 2540258, FriendID: 8, DisplayOrder: 3}, // Travis Kelce 1
				{ID: 7, SourceID: 2552475, FriendID: 8, DisplayOrder: 1}, // Todd Gurley 1
				{ID: 6, SourceID: 2495454, FriendID: 8, DisplayOrder: 2}, // Julio Jones 3
			},
			playersJSON: `{"players":[
				{"id":"2495454","name":"Julio Jones","position":"WR","stats":{"22":"3"}},
				{"id":"2540258","name":"Travis Kelce","position":"TE","stats":{"22":"1"}},
				{"id":"2552475","name":"Todd Gurley","position":"RB","stats":{"15":"1"}}]}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeNflMisc,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 8, Name: "Dave", Score: 4,
						PlayerScores: []PlayerScore{
							{ID: 7, Name: "Todd Gurley", Score: 1, DisplayOrder: 1, SourceID: 2552475},
							{ID: 6, Name: "Julio Jones", Score: 3, DisplayOrder: 2, SourceID: 2495454},
							{ID: 1, Name: "Travis Kelce", Score: 1, DisplayOrder: 3, SourceID: 2540258},
						},
					},
				},
			},
		},
	}
	for i, test := range requestScoreCategoryTests {
		jsonFunc := func(uri string) string {
			return test.playersJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		nflPlayerRequestor := nflPlayerRequestor{requestor: r}
		got, err := nflPlayerRequestor.requestScoreCategory(test.pt, 2019, test.friends, test.players)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: wanted error but did not get one", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		case !reflect.DeepEqual(test.want, got):
			t.Errorf("Test %v: Not equal:\nWanted: %v\nGot:    %v", i, test.want, got)
		}
	}
}

func TestNflPlayerPlayerSearchResults(t *testing.T) {
	playerSearchResultsTests := []struct {
		pt               db.PlayerType
		playerNamePrefix string
		playersJSON      string
		wantErr          bool
		want             []PlayerSearchResult
	}{
		{
			playerNamePrefix: "cam",
			wantErr:          true, // no playersJSON
		},
		{
			pt:               db.PlayerTypeNflQB,
			playerNamePrefix: "russell",
			playersJSON: `{"players":[
				{"id":"2541944","name":"Russell Shepard","position":"WR","teamAbbr":"NYG","stats":{"1":"1"}},
				{"id":"2562717","name":"Dontavius Russell","position":"DL","teamAbbr":"JAX","stats":{"1":"1"}},
				{"id":"2532975","name":"Russell Wilson","position":"QB","teamAbbr":"SEA","stats":{"1":"2","6":"5"}}]}`,
			want: []PlayerSearchResult{
				{Name: "Russell Wilson", Details: "Team: SEA, Position: QB", SourceID: 2532975},
			},
		},
	}
	for i, test := range playerSearchResultsTests {
		jsonFunc := func(uri string) string {
			return test.playersJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		nflPlayerRequestor := nflPlayerRequestor{requestor: r}
		got, err := nflPlayerRequestor.search(test.pt, 2019, test.playerNamePrefix, true)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: wanted error but did not get one", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		case !reflect.DeepEqual(test.want, got):
			t.Errorf("Test %v: Not equal:\nWanted: %v\nGot:    %v", i, test.want, got)
		}
	}
}
