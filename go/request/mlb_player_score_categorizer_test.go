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
		{ // simple case
			playerStatsJSON: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`,
			playerType:      db.PlayerTypeMlbHitter,
			want:            39,
		},
		{ // unknown player type.  Negative number is invalid score
			playerStatsJSON: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`,
			playerType:      0,
			want:            -1,
		},
		{ // Luis Severino did not play in[most of] 2019, so the score should be 0 [midway through the season]
			playerStatsJSON: `{"stats":[]}`,
			playerType:      db.PlayerTypeMlbPitcher,
			want:            0,
		},
		{ // Edwin Encarnacion played for multiple teams in 2019, so the last Stat's score should be returned (multiplying homeRuns from first team by 10 to ensure this)
			playerStatsJSON: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":210}},{"stat":{"homeRuns":9}},{"stat":{"homeRuns":30}}]}]}`,
			playerType:      db.PlayerTypeMlbHitter,
			want:            30,
		},
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

func TestMlbPlayerRequestScoreCategory(t *testing.T) {
	RequestScoreCategoryTests := []struct {
		pt               db.PlayerType
		friends          []db.Friend
		players          []db.Player
		playerNamesJSON  string
		playerStatsJSONs map[db.ID]string
		wantErr          bool
		want             ScoreCategory
	}{
		{
			pt: db.PlayerTypeMlbHitter,
			friends: []db.Friend{
				{ID: "1", DisplayOrder: 2, Name: "Bobby"},
				{ID: "2", DisplayOrder: 1, Name: "Charles"},
			},
			players: []db.Player{
				{ID: "1", SourceID: 547180, FriendID: "1", DisplayOrder: 1}, // Bryce Harper 31
				{ID: "2", SourceID: 545361, FriendID: "1", DisplayOrder: 3}, // Mike Trout 45
				{ID: "3", SourceID: 547180, FriendID: "2", DisplayOrder: 1}, // Bryce Harper 31
				{ID: "4", SourceID: 429665, FriendID: "1", DisplayOrder: 2}, // Edwin Encarnacion 34
			},
			playerNamesJSON: `{"People":[
				{"id":545361,"fullName":"Mike Trout"},
				{"id":429665,"fullName":"Edwin Encarnacion"},
				{"id":547180,"fullName":"Bryce Harper"}]}`,
			playerStatsJSONs: map[db.ID]string{
				"547180": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":31}}]}]}`,
				"545361": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":45}}]}]}`,
				"429665": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":34}}]}]}`,
			},
			want: ScoreCategory{
				PlayerType: db.PlayerTypeMlbHitter,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: "2", Name: "Charles", Score: 31,
						PlayerScores: []PlayerScore{
							{ID: "3", Name: "Bryce Harper", Score: 31, DisplayOrder: 1, SourceID: 547180}},
					},
					{
						DisplayOrder: 2, ID: "1", Name: "Bobby", Score: 79, // only sum top two scores
						PlayerScores: []PlayerScore{
							{ID: "1", Name: "Bryce Harper", Score: 31, DisplayOrder: 1, SourceID: 547180},
							{ID: "4", Name: "Edwin Encarnacion", Score: 34, DisplayOrder: 2, SourceID: 429665},
							{ID: "2", Name: "Mike Trout", Score: 45, DisplayOrder: 3, SourceID: 545361},
						},
					},
				},
			},
		},
		{
			pt:               db.PlayerTypeMlbPitcher,
			friends:          []db.Friend{{ID: "8", DisplayOrder: 1, Name: "Brandon"}},
			players:          []db.Player{{ID: "7", SourceID: 605483, FriendID: "8", DisplayOrder: 1}}, // Blake Snell 6
			playerNamesJSON:  `{"People":[{"id":605483,"fullName":"Blake Snell"}]}`,
			playerStatsJSONs: map[db.ID]string{"605483": `{"stats":[{"group":{"displayName":"pitching"},"splits":[{"stat":{"wins":6}}]}]}`},
			want: ScoreCategory{
				PlayerType: db.PlayerTypeMlbPitcher,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: "8", Name: "Brandon", Score: 6,
						PlayerScores: []PlayerScore{
							{ID: "7", Name: "Blake Snell", Score: 6, DisplayOrder: 1, SourceID: 605483}},
					},
				},
			},
		},
		{ // no players
			pt:      db.PlayerTypeMlbPitcher,
			friends: []db.Friend{{ID: "8", DisplayOrder: 1, Name: "Brandon"}},
			want: ScoreCategory{
				PlayerType: db.PlayerTypeMlbPitcher,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: "8", Name: "Brandon", Score: 0,
						PlayerScores: []PlayerScore{},
					},
				},
			},
		},
		{
			pt:               db.PlayerTypeMlbHitter,
			players:          []db.Player{{ID: "7", SourceID: 592450, FriendID: "9", DisplayOrder: 1}}, // Aaron Judge 24
			playerNamesJSON:  `{"People":[]}`,
			playerStatsJSONs: map[db.ID]string{"592450": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":24}}]}]}`},
			wantErr:          true, // incorrect number of names
		},
		{
			pt:               db.PlayerTypeNflQB,
			players:          []db.Player{{ID: "9", SourceID: 2532975, FriendID: "6", DisplayOrder: 1}}, // Russell Wilson 0
			playerNamesJSON:  `{"People":[{"id":2532975,"fullName":"Russell Wilson"}]}`,
			playerStatsJSONs: map[db.ID]string{"2532975": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":0}}]}]}`},
			wantErr:          true, // incorrect playerType for MlbPlayerStats.getStat(pt)
		},
		{
			pt:               db.PlayerTypeMlbHitter,
			players:          []db.Player{{ID: "7", SourceID: 592450, FriendID: "9", DisplayOrder: 1}}, // Aaron Judge 24
			playerStatsJSONs: map[db.ID]string{"2532975": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":24}}]}]}`},
			wantErr:          true, // no playerNamesJSON
		},
		{
			pt:               db.PlayerTypeMlbHitter,
			players:          []db.Player{{ID: "7", SourceID: 592450, FriendID: "9", DisplayOrder: 1}}, // Aaron Judge 24
			playerNamesJSON:  `Aaron Judge`,
			playerStatsJSONs: map[db.ID]string{"2532975": `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":24}}]}]}`},
			wantErr:          true, // bad playerNamesJSON
		},
		{
			pt:              db.PlayerTypeMlbHitter,
			players:         []db.Player{{ID: "7", SourceID: 592450, FriendID: "9", DisplayOrder: 1}}, // Aaron Judge 24
			playerNamesJSON: `{"People":[{"id":592450,"fullName":"Aaron Judge"}]}`,
			wantErr:         true, // no playerStatsJSON
		},
		{
			pt:               db.PlayerTypeMlbPitcher,
			friends:          []db.Friend{{ID: "4", DisplayOrder: 1, Name: "Cameron"}},
			players:          []db.Player{{ID: "2", SourceID: 622663, FriendID: "4", DisplayOrder: 1}}, // Luis Severino 0
			playerNamesJSON:  `{"People":[{"id":622663,"fullName":"Luis Severino"}]}`,
			playerStatsJSONs: map[db.ID]string{"622663": `{"stats":[]}`}, // no stats
			want: ScoreCategory{
				PlayerType: db.PlayerTypeMlbPitcher,
				FriendScores: []FriendScore{
					{
						DisplayOrder: 1, ID: "4", Name: "Cameron", Score: 0,
						PlayerScores: []PlayerScore{
							{ID: "2", Name: "Luis Severino", Score: 0, DisplayOrder: 1, SourceID: 622663}},
					},
				},
			},
		},
	}
	for i, test := range RequestScoreCategoryTests {
		jsonFunc := func(uri string) string {
			if strings.Contains(uri, "/people?") {
				return test.playerNamesJSON
			}
			for playerSourceID, playerStatsJSON := range test.playerStatsJSONs {
				if strings.Contains(uri, fmt.Sprintf("%v/stats?", playerSourceID)) {
					return playerStatsJSON
				}
			}
			return "" // will cause json unmarshal error
		}
		r := newMockHTTPRequester(jsonFunc)
		mlbPlayerRequester := mlbPlayerRequester{requester: r}
		got, err := mlbPlayerRequester.RequestScoreCategory(test.pt, db.PlayerTypeInfo{}, 2019, test.friends, test.players)
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
