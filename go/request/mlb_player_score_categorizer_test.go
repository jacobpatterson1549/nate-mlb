package request

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type mlbPlayerStatsTest struct {
	playerStatsJSON string
	playerType      db.PlayerType
	want            int
}

var mlbPlayerStatsTests = []mlbPlayerStatsTest{
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, db.PlayerTypeHitter, 39}, // simple case
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, 0, -1},                   // unknown player type.  Negative number is invalid score
	{`{"stats":[]}`, db.PlayerTypePitcher, 0}, // Luis Severino did not play in 2019, so the score should be 0
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":21}},{"stat":{"homeRuns":9}},{"stat":{"homeRuns":30}}]}]}`, db.PlayerTypeHitter, 30}, // Edwin Encarnacion played for multiple teams in 2019, so the last Stat's score should be returned
}

func TestLastStatScore(t *testing.T) {
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

type mockHTTPClient struct {
	JSONFunc func(urlPath string) string
}

func (m mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(m.JSONFunc(r.URL.Path))),
	}, nil
}

func TestRequestScoreCategoryHitters(t *testing.T) {
	friends := []db.Friend{
		{
			ID:           1,
			DisplayOrder: 2,
			Name:         "Bobby",
		},
		{
			ID:           2,
			DisplayOrder: 1,
			Name:         "Charles",
		},
	}
	players := map[db.PlayerType][]db.Player{
		db.PlayerTypeHitter: []db.Player{
			{
				ID:           1,
				SourceID:     547180, // Bryce Harper 31
				FriendID:     1,
				DisplayOrder: 1,
			},
			{
				ID:           2,
				SourceID:     545361, // Mike Trout 45
				FriendID:     1,
				DisplayOrder: 3,
			},
			{
				ID:           3,
				SourceID:     547180, // Bryce Harper 31
				FriendID:     2,
				DisplayOrder: 1,
			},
			{
				ID:           4,
				SourceID:     429665, // Edwin Encarnacion 34
				FriendID:     1,
				DisplayOrder: 2,
			},
		},
	}
	year := 2019
	fpi := FriendPlayerInfo{
		Friends: friends,
		Players: players,
		Year:    year,
	}
	jsonFunc := func(urlPath string) string {
		switch {
		case strings.HasSuffix(urlPath, "/people"):
			return `{"People": [
				{ "id": 545361, "fullName": "Mike Trout" },
				{ "id": 429665, "fullName": "Edwin Encarnacion" },
				{ "id": 547180, "fullName": "Bryce Harper" }]}`
		case strings.HasSuffix(urlPath, "547180/stats"): // Bryce Harper
			return `{ "stats" : [ { "group" : { "displayName" : "hitting" },
				"splits" : [ { "stat" : { "homeRuns" : 31 } } ] } ] }`
		case strings.HasSuffix(urlPath, "545361/stats"): // Mike Trout
			return `{ "stats" : [ { "group" : { "displayName" : "hitting" },
				"splits" : [ { "stat" : { "homeRuns" : 45 } } ] } ] }`
		case strings.HasSuffix(urlPath, "429665/stats"): // Edwin Encarnacion
			return `{ "stats" : [ { "group" : { "displayName" : "hitting" },
				"splits" : [ { "stat" : { "homeRuns" : 34 } } ] } ] }`
		}
		return "null"
	}
	r := httpRequestor{
		cache:          newCache(0),
		httpClient:     mockHTTPClient{JSONFunc: jsonFunc},
		logRequestUrls: true,
	}
	mlbPlayerRequestor := mlbPlayerRequestor{requestor: &r}

	want := ScoreCategory{
		PlayerType: db.PlayerTypeHitter,
		FriendScores: []FriendScore{
			{
				ID:           2,
				Name:         "Charles",
				Score:        31,
				DisplayOrder: 1,
				PlayerScores: []PlayerScore{
					{
						ID:           3,
						Name:         "Bryce Harper",
						Score:        31,
						DisplayOrder: 1,
						SourceID:     547180,
					},
				},
			},
			{
				ID:           1,
				Name:         "Bobby",
				Score:        79, // only sum top two scores
				DisplayOrder: 2,
				PlayerScores: []PlayerScore{
					{
						ID:           1,
						Name:         "Bryce Harper",
						Score:        31,
						DisplayOrder: 1,
						SourceID:     547180,
					},
					{
						ID:           4,
						Name:         "Edwin Encarnacion",
						Score:        34,
						DisplayOrder: 2,
						SourceID:     429665,
					},
					{
						ID:           2,
						Name:         "Mike Trout",
						Score:        45,
						DisplayOrder: 3,
						SourceID:     545361,
					},
				},
			},
		},
	}
	got, err := mlbPlayerRequestor.RequestScoreCategory(fpi, db.PlayerTypeHitter)

	switch {
	case err != nil:
		t.Error(err)
	case !reflect.DeepEqual(want, got): // TODO: use reflect.DeepEqual in export_test, cache_test, mlb_player_searcher_test
		t.Errorf("Not equal:\nWanted: %v\nGot:    %v", want, got)
	}
}
