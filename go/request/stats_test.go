package request

import (
	"testing"
)

func TestGetFriendScore(t *testing.T) {
	getFriendScoreTests := []struct {
		playerScores              []PlayerScore
		onlySumTopTwoPlayerScores bool
		want                      int
	}{
		{
			// basic sum
			playerScores: []PlayerScore{
				{Score: 1},
				{Score: 2},
				{Score: 3},
			},
			onlySumTopTwoPlayerScores: false,
			want:                      6,
		},
		{
			// basic sum top two
			playerScores: []PlayerScore{
				{Score: 1},
				{Score: 2},
				{Score: 3},
			},
			onlySumTopTwoPlayerScores: true,
			want:                      5,
		},
		{
			// one playerScore
			playerScores: []PlayerScore{
				{Score: 44},
			},
			onlySumTopTwoPlayerScores: true,
			want:                      44,
		},
		{
			// no playerScores
			onlySumTopTwoPlayerScores: true,
			want:                      0,
		},
	}
	for i, test := range getFriendScoreTests {
		got := getFriendScore(test.playerScores, test.onlySumTopTwoPlayerScores)
		if test.want != got {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
	}
}
