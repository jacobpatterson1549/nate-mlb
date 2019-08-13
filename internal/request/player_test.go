package request

import (
	"encoding/json"
	"nate-mlb/internal/db"
	"testing"
)

type playerStatsScoreTest struct {
	playerStatsJSON string
	playerType      db.PlayerType
	want            int
}

var playerStatsScoreTests = []playerStatsScoreTest{
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, db.PlayerTypeHitter, 39}, // simple case
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":39}}]}]}`, db.PlayerType(0), -1},    // unknown player type.  Negative number is invalid score
	{`{"stats":[]}`, db.PlayerTypePitcher, 0}, // Luis Severino did not play in 2019, so the score should be 0
	{`{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":21}},{"stat":{"homeRuns":9}},{"stat":{"homeRuns":30}}]}]}`, db.PlayerTypeHitter, 30}, // Edwin Encarnacion played for multiple teams in 2019, so the last Stat's score should be returned
}

func TestLastStatScore(t *testing.T) {
	for i, test := range playerStatsScoreTests {
		var playerStats PlayerStats
		err := json.Unmarshal([]byte(test.playerStatsJSON), &playerStats)
		if err != nil {
			t.Errorf("Test %v: %v", i, err)
		}
		got, err := playerStats.getScore(test.playerType)
		if got != test.want {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
		if err != nil && got >= 0 {
			t.Errorf("Test %v: wanted %v, but got ERROR: %v", i, test.want, err)
		}
	}
}
