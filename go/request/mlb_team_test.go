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
			players: []db.Player{{ID: 8, SourceID: 136, FriendID: 3, DisplayOrder: 1}},
			wantErr: true, // no teamsJSON
		},
	}
	for i, test := range requestScoreCategoryTests {
		jsonFunc := func(urlPath string) string {
			return test.teamsJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		mlbPlayerRequestor := mlbTeamRequestor{requestor: r}
		got, err := mlbPlayerRequestor.RequestScoreCategory(db.PlayerTypeMlbTeam, 2001, test.friends, test.players)
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
