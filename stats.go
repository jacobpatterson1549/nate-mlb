package main

func getStats(friendPlayerInfo FriendPlayerInfo) []ScoreCategory {
	return nil
}

// ScoreCategory  contain the FriendScores for each PlayerType
type ScoreCategory struct {
	name         string
	friendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	friendName   string
	playerScores []PlayerScore
	score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	playerName string
	score      int
}
