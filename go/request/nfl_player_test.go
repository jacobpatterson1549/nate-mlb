package request

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestNflPlayerRequestScoreCategory(t *testing.T) {
	RequestScoreCategoryTests := []struct {
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
			playersJSON: `{"games":{"102020":{"players":{
				"2532975":{"playerId":"2532975","name":"Russell Wilson","position":"QB","stats":{"season":{"2018":{"1":"16","6":"35"}}}}
				}}}}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeNflQB,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 2, Name: "Carl", Score: 35,
						PlayerScores: []PlayerScore{
							{ID: 3, Name: "Russell Wilson", Score: 35, DisplayOrder: 1, SourceID: 2532975},
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
			playersJSON: `{"games":{"102020":{"players":{
				"2495454":{"playerId":"2495454","name":"Julio Jones","position":"WR","stats":{"season":{"2018":{"22":"8"}}}},
				"2540258":{"playerId":"2540258","name":"Travis Kelce","position":"TE","stats":{"season":{"2018":{"22":"10"}}}},
				"2552475":{"playerId":"2552475","name":"Todd Gurley","position":"RB","stats":{"season":{"2018":{"15":"17"}}}}
				}}}}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeNflMisc,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 8, Name: "Dave", Score: 27,
						PlayerScores: []PlayerScore{
							{ID: 7, Name: "Todd Gurley", Score: 17, DisplayOrder: 1, SourceID: 2552475},
							{ID: 6, Name: "Julio Jones", Score: 8, DisplayOrder: 2, SourceID: 2495454},
							{ID: 1, Name: "Travis Kelce", Score: 10, DisplayOrder: 3, SourceID: 2540258},
						},
					},
				},
			},
		},
	}
	for i, test := range RequestScoreCategoryTests {
		jsonFunc := func(uri string) string {
			return test.playersJSON
		}
		r := newMockHTTPRequester(jsonFunc)
		nflPlayerRequester := nflPlayerRequester{requester: r}
		got, err := nflPlayerRequester.RequestScoreCategory(test.pt, db.PlayerTypeInfo{}, 2019, test.friends, test.players)
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
			playersJSON: `{"games":{"102020":{"players":{
				"2541944":{"playerId":"2541944","name":"Russell Shepard","position":"WR","nflTeamAbbr":"NYG"},
				"2562717":{"playerId":"2562717","name":"Dontavius Russell","position":"DL","nflTeamAbbr":"JAX"},
				"2532975":{"playerId":"2532975","name":"Russell Wilson","position":"QB","nflTeamAbbr":"SEA"}
				}}}}`,
			want: []PlayerSearchResult{
				{Name: "Russell Wilson", Details: "Team: SEA, Position: QB", SourceID: 2532975},
			},
		},
		{
			playersJSON: `{}`,
		},
	}
	for i, test := range playerSearchResultsTests {
		jsonFunc := func(uri string) string {
			return test.playersJSON
		}
		r := newMockHTTPRequester(jsonFunc)
		nflPlayerRequester := nflPlayerRequester{requester: r}
		got, err := nflPlayerRequester.Search(test.pt, 2019, test.playerNamePrefix, true)
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

func TestNflPlayerGetStats(t *testing.T) {
	nflPlayerStatsTests := []struct {
		stats   map[string]json.RawMessage
		want    NflPlayerStats
		wantErr bool
	}{
		{
			stats: map[string]json.RawMessage{
				"season": json.RawMessage(`{"2018":{"6":"35"}}`),
			},
			want: NflPlayerStats{
				PassingTD: 35,
			},
		},
		{
			stats: map[string]json.RawMessage{ // bad key
				"week": json.RawMessage(`{"2018":{"6":"35"}}`),
			},
			wantErr: true,
		},
		{
			stats: map[string]json.RawMessage{ // bad json
				"season": json.RawMessage(`{"2018":{"PassingTD":"35"}}`),
			},
			wantErr: true,
		},
		{
			stats: map[string]json.RawMessage{ // any of the two could be picked, but should not crash
				"season": json.RawMessage(`{"2018":{},"2019":{}}`),
			},
			want: NflPlayerStats{},
		},
	}
	for i, test := range nflPlayerStatsTests {
		nflPlayer := NflPlayer{
			Stats: test.stats,
		}
		got, err := nflPlayer.stats()
		switch {
		case err != nil:
			if !test.wantErr {
				t.Errorf("Test %v: unexpected error: %v", i, err)
			}
		case test.want != got:
			t.Errorf("Test %v:\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}
