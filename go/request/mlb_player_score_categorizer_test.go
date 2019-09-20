package request

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestLastStatScore(t *testing.T) {
	var mlbPlayerStatsTests = []struct {
		playerStatsJSON string
		playerType      db.PlayerType
		want            int
	}{
		{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, db.PlayerTypeHitter, 39}, // simple case
		{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, 0, -1},                   // unknown player type.  Negative number is invalid score
		{`{"stats":[]}`, db.PlayerTypePitcher, 0}, // Luis Severino did not play in 2019, so the score should be 0
		{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":21}},{"stat":{"homeRuns":9}},{"stat":{"homeRuns":30}}]}]}`, db.PlayerTypeHitter, 30}, // Edwin Encarnacion played for multiple teams in 2019, so the last Stat's score should be returned
	}
	for i, test := range mlbPlayerStatsTests {
		var mlbPlayerStats MlbPlayerStats
		err := json.Unmarshal([]byte(test.playerStatsJSON), &mlbPlayerStats)
		if err != nil {
			t.Errorf("Test %v: %v", i, err)
		}
		got, err := mlbPlayerStats.getStat(test.playerType)
		if got != test.want {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
		if err != nil && got >= 0 {
			t.Errorf("Test %v: wanted %v, but got ERROR: %v", i, test.want, err)
		}
	}
}

func TestMlbRequestScoreCategoryHitters(t *testing.T) {
	var requestScoreCategoryTests = []struct {
		pt               db.PlayerType
		year             int
		friends          []db.Friend
		players          []db.Player
		playerNamesJSON  string
		playerStatsJSONs map[db.ID]string
		wantErr          bool
		want             ScoreCategory
	}{
		{
			pt:   db.PlayerTypeHitter,
			year: 2019,
			friends: []db.Friend{
				{ID: 1, DisplayOrder: 2, Name: "Bobby"},
				{ID: 2, DisplayOrder: 1, Name: "Charles"},
			},
			players: []db.Player{
				{ID: 1, SourceID: 547180, FriendID: 1, DisplayOrder: 1}, // Bryce Harper 31
				{ID: 2, SourceID: 545361, FriendID: 1, DisplayOrder: 3}, // Mike Trout 45
				{ID: 3, SourceID: 547180, FriendID: 2, DisplayOrder: 1}, // Bryce Harper 31
				{ID: 4, SourceID: 429665, FriendID: 1, DisplayOrder: 2}, // Edwin Encarnacion 34
			},
			playerNamesJSON: `{"People":[
				{"id": 545361,"fullName":"Mike Trout"},
				{"id": 429665,"fullName":"Edwin Encarnacion"},
				{"id": 547180,"fullName":"Bryce Harper"}]}`,
			playerStatsJSONs: map[db.ID]string{
				547180: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":31}}]}]}`,
				545361: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":45}}]}]}`,
				429665: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":34}}]}]}`,
			},
			want: ScoreCategory{
				PlayerType: db.PlayerTypeHitter,
				FriendScores: []FriendScore{
					{
						ID:           2,
						Name:         "Charles",
						Score:        31,
						DisplayOrder: 1,
						PlayerScores: []PlayerScore{
							{ID: 3, Name: "Bryce Harper", Score: 31, DisplayOrder: 1, SourceID: 547180}},
					},
					{
						ID:           1,
						Name:         "Bobby",
						Score:        79, // only sum top two scores
						DisplayOrder: 2,
						PlayerScores: []PlayerScore{
							{ID: 1, Name: "Bryce Harper", Score: 31, DisplayOrder: 1, SourceID: 547180},
							{ID: 4, Name: "Edwin Encarnacion", Score: 34, DisplayOrder: 2, SourceID: 429665},
							{ID: 2, Name: "Mike Trout", Score: 45, DisplayOrder: 3, SourceID: 545361},
						},
					},
				},
			},
		},
	}
	for i, test := range requestScoreCategoryTests {
		jsonFunc := func(urlPath string) string {
			if strings.HasSuffix(urlPath, "/people") {
				return test.playerNamesJSON
			}
			for playerSourceID, playerStatsJSON := range test.playerStatsJSONs {
				urlSuffix := fmt.Sprintf("%d/stats", playerSourceID)
				if strings.HasSuffix(urlPath, urlSuffix) {
					return playerStatsJSON
				}
			}
			return "null"
		}
		r := newMockHTTPRequestor(jsonFunc)
		mlbPlayerRequestor := mlbPlayerRequestor{requestor: r}
		got, err := mlbPlayerRequestor.RequestScoreCategory(test.pt, test.year, test.friends, test.players)
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
