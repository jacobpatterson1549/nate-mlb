package request

import (
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestMlbTeamRequestScoreCategory(t *testing.T) {
	requestScoreCategoryTests := []struct {
		friends   []db.Friend
		players   []db.Player
		teamsJSON string
		wantErr   bool
		want      ScoreCategory
	}{
		{
			friends: []db.Friend{{ID: 3, DisplayOrder: 1, Name: "Elias"}},
			players: []db.Player{
				{ID: 5, SourceID: 133, FriendID: 3, DisplayOrder: 2}, // Oakland Athletics 102
				{ID: 8, SourceID: 136, FriendID: 3, DisplayOrder: 1}, // Seattle Mariners 116
				{ID: 9, SourceID: 112, FriendID: 3, DisplayOrder: 3}, // Chicago Cubs 88

			},
			teamsJSON: `{"records":[
				{"teamRecords":[
						{"team":{"id":133,"name":"Oakland Athletics"},"wins":102},
						{"team":{"id":136,"name":"Seattle Mariners"},"wins":116}]},
				{"teamRecords":[
						{"team":{"id":112,"name":"Chicago Cubs"},"wins":88}]}]}`,
			want: ScoreCategory{
				PlayerType: db.PlayerTypeMlbTeam,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: 3, Name: "Elias", Score: 306,
						PlayerScores: []PlayerScore{
							{ID: 8, Name: "Seattle Mariners", Score: 116, DisplayOrder: 1, SourceID: 136},
							{ID: 5, Name: "Oakland Athletics", Score: 102, DisplayOrder: 2, SourceID: 133},
							{ID: 9, Name: "Chicago Cubs", Score: 88, DisplayOrder: 3, SourceID: 112},
						},
					},
				},
			},
		},
		{
			wantErr: true, // no teamsJSON
		},
	}
	for i, test := range requestScoreCategoryTests {
		jsonFunc := func(urlPath string) string {
			return test.teamsJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		mlbTeamRequestor := mlbTeamRequestor{requestor: r}
		got, err := mlbTeamRequestor.RequestScoreCategory(db.PlayerTypeMlbTeam, 2001, test.friends, test.players)
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

func TestMlbTeamPlayerSearchResults(t *testing.T) {
	playerSearchResultsTests := []struct {
		playerNamePrefix string
		teamsJSON        string
		wantErr          bool
		want             []PlayerSearchResult
	}{
		{
			playerNamePrefix: "c",
			teamsJSON: `{"records":[
				{"teamRecords":[
						{"team":{"id":133,"name":"Oakland Athletics"},"wins":102,"losses":60},
						{"team":{"id":136,"name":"Seattle Mariners"},"wins":116,"losses":46}]},
				{"teamRecords":[
						{"team":{"id":112,"name":"Chicago Cubs"},"wins":88,"losses":74}]}]}`,
			want: []PlayerSearchResult{
				{Name: "Oakland Athletics", Details: "102 - 60 Record", SourceID: 133},
				{Name: "Chicago Cubs", Details: "88 - 74 Record", SourceID: 112},
			},
		},
		{
			playerNamePrefix: "Sox",
			wantErr:          true, // no teamsJSON
		},
	}
	for i, test := range playerSearchResultsTests {
		jsonFunc := func(urlPath string) string {
			return test.teamsJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		mlbTeamRequestor := mlbTeamRequestor{requestor: r}
		got, err := mlbTeamRequestor.PlayerSearchResults(db.PlayerTypeMlbTeam, 2019, test.playerNamePrefix, true)
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
