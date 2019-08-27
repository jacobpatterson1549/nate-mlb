package request

import (
	"testing"
)

type populateFriendScoreTest struct {
	friendScore               FriendScore
	onlySumTopTwoPlayerScores bool
	want                      int
}

var populateFriendScoreTests = []populateFriendScoreTest{
	{
		// basic sum
		friendScore: FriendScore{
			PlayerScores: []PlayerScore{
				{Score: 1},
				{Score: 2},
				{Score: 3},
			},
		},
		onlySumTopTwoPlayerScores: false,
		want:                      6,
	},
	{
		// basic sum top two
		friendScore: FriendScore{
			PlayerScores: []PlayerScore{
				{Score: 1},
				{Score: 2},
				{Score: 3},
			},
		},
		onlySumTopTwoPlayerScores: true,
		want:                      5,
	},
	{
		// one playerScore
		friendScore: FriendScore{
			PlayerScores: []PlayerScore{
				{Score: 44},
			},
		},
		onlySumTopTwoPlayerScores: true,
		want:                      44,
	},
	{
		// no playerScores
		friendScore: FriendScore{
			Score: 17, // [should be overwritten]
		},
		onlySumTopTwoPlayerScores: true,
		want:                      0,
	},
}

func TestFriendScorePopulateScore(t *testing.T) {
	for i, test := range populateFriendScoreTests {
		test.friendScore.populateScore(test.onlySumTopTwoPlayerScores)
		got := test.friendScore.Score

		if test.want != got {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
	}
}
