package request

import (
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestNflTeamRequestScoreCategory(t *testing.T) {
	requestScoreCategoryTests := []struct {
		friends   []db.Friend
		players   []db.Player
		teamsJSON string
		wantErr   bool
		want      ScoreCategory
	}{
		{
			wantErr: true, // no teamsJSON
		},
		{
			teamsJSON: `{"nflTeams":{"30":{"fullName":"Seattle Seahawks","record":"all-0-0"}}}`,
			wantErr:   true, // bad wins
		},
		{
			friends: []db.Friend{{ID: 7, DisplayOrder: 1, Name: "Anthony"}},
			players: []db.Player{
				{ID: 4, SourceID: 30, FriendID: 7, DisplayOrder: 1}, // Seattle Seahawks 10
				{ID: 1, SourceID: 20, FriendID: 7, DisplayOrder: 3}, // Minnesota Vikings 8
				{ID: 3, SourceID: 29, FriendID: 7, DisplayOrder: 2}, // San Francisco 49er 4
			},
			teamsJSON: `{"nflTeams":{
				"20":{"fullName":"Minnesota Vikings","record":"8-7-1"},
				"29":{"fullName":"San Francisco 49ers","record":"4-12-0"},
				"30":{"fullName":"Seattle Seahawks","record":"10-6-0"}}}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeNflTeam,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 7, Name: "Anthony", Score: 22,
						PlayerScores: []PlayerScore{
							{ID: 4, Name: "Seattle Seahawks", Score: 10, DisplayOrder: 1, SourceID: 30},
							{ID: 3, Name: "San Francisco 49ers", Score: 4, DisplayOrder: 2, SourceID: 29},
							{ID: 1, Name: "Minnesota Vikings", Score: 8, DisplayOrder: 3, SourceID: 20},
						},
					},
				},
			},
		},
	}
	for i, test := range requestScoreCategoryTests {
		jsonFunc := func(uri string) string {
			return test.teamsJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		nflTeamRequestor := nflTeamRequestor{requestor: r}
		got, err := nflTeamRequestor.requestScoreCategory(db.PlayerTypeNflTeam, 2019, test.friends, test.players)
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

func TestNflTeamPlayerSearchResults(t *testing.T) {
	playerSearchResultsTests := []struct {
		playerNamePrefix string
		teamsJSON        string
		wantErr          bool
		want             []PlayerSearchResult
	}{
		{
			playerNamePrefix: "cowboys",
			wantErr:          true, // no teamsJSON
		},
		{
			playerNamePrefix: "Bay",
			teamsJSON: `{"nflTeams":{
				"2":{"nflTeamId":"2","fullName":"Baltimore Ravens","record":"2-0-0"},
				"31":{"nflTeamId":"31","fullName":"Tampa Bay Buccaneers","record":"1-1-0"},
				"11":{"nflTeamId":"11","fullName":"Green Bay Packers","record":"2-0-0"}}}`,
			want: []PlayerSearchResult{
				{Name: "Green Bay Packers", Details: "2-0-0 Record", SourceID: 11},
				{Name: "Tampa Bay Buccaneers", Details: "1-1-0 Record", SourceID: 31},
			},
		},
	}
	for i, test := range playerSearchResultsTests {
		jsonFunc := func(uri string) string {
			return test.teamsJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		nflTeamRequestor := nflTeamRequestor{requestor: r}
		got, err := nflTeamRequestor.search(db.PlayerTypeMlbTeam, 2019, test.playerNamePrefix, true)
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

func TestNflTeamWins(t *testing.T) {
	nflTeamWinsTests := []struct {
		nflTeam   NflTeam
		want      int
		wantError bool
	}{
		{
			nflTeam: NflTeam{Record: "7-9-0"},
			want:    7,
		},
		{
			nflTeam: NflTeam{Record: "6-9-1"},
			want:    6,
		},
		{
			nflTeam: NflTeam{Record: "16"},
			want:    16,
		},
		{
			nflTeam:   NflTeam{Record: ""},
			wantError: true,
		},
		{
			nflTeam:   NflTeam{Record: "eight-8-0"},
			wantError: true,
		},
		{
			nflTeam:   NflTeam{Record: "-4-12-0"},
			wantError: true,
		},
		{
			nflTeam:   NflTeam{Record: "four and ten"},
			wantError: true,
		},
	}
	for i, test := range nflTeamWinsTests {
		got, err := test.nflTeam.wins()
		switch {
		case test.wantError:
			if err == nil {
				t.Errorf("Test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("Test %v: %v", i, err)
		case test.want != got:
			t.Errorf("Test %v: wanted %v, got %v", i, test.want, got)
		}
	}
}
