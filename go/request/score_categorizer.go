package request

import (
	"sort"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// ScoreCategorizer requests data for and creates a ScoreCategory for the FriendPlayerInfo
	scoreCategorizer interface {
		requestScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error)
	}

	// ScoreCategory contain the FriendScores for each PlayerType
	ScoreCategory struct {
		Name         string
		Description  string
		PlayerType   db.PlayerType // Used as an int on the website
		FriendScores []FriendScore
	}

	// FriendScore contain the scores for a Friend for a PlayerType
	FriendScore struct {
		ID           db.ID
		Name         string
		ScoreType    string
		Score        int
		DisplayOrder int
		PlayerScores []PlayerScore
	}

	// PlayerScore is the score for a particular Player
	PlayerScore struct {
		ID           db.ID
		Name         string
		Score        int
		DisplayOrder int
		SourceID     db.SourceID
	}

	playerName struct {
		sourceID db.SourceID
		name     string
	}

	playerStat struct {
		sourceID db.SourceID
		stat     int
	}

	nameScore struct {
		name  string
		score int
	}
)

func newScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, friends []db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) ScoreCategory {
	return ScoreCategory{
		Name:         ptInfo.Name,
		PlayerType:   pt,
		Description:  ptInfo.Description,
		FriendScores: newFriendScores(ptInfo.ScoreType, friends, players, playerNameScores, onlySumTopTwoPlayerScores),
	}
}

func newFriendScores(scoreType string, friends []db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) []FriendScore {
	friendPlayers := make(map[db.ID][]db.Player, len(players))
	for _, player := range players {
		friendPlayers[player.FriendID] = append(friendPlayers[player.FriendID], player)
	}
	friendScores := make([]FriendScore, len(friends))
	for i, friend := range friends {
		friendScores[i] = newFriendScore(scoreType, friend, friendPlayers[friend.ID], playerNameScores, onlySumTopTwoPlayerScores)
	}
	displayOrder := func(i int) int { return friendScores[i].DisplayOrder }
	sort.Slice(friendScores, func(i, j int) bool {
		return displayOrder(i) < displayOrder(j)
	})
	return friendScores
}

func newFriendScore(scoreType string, friend db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) FriendScore {
	playerScores := newPlayerScores(players, playerNameScores)
	return FriendScore{
		ID:           friend.ID,
		Name:         friend.Name,
		ScoreType:    scoreType,
		Score:        getFriendScore(playerScores, onlySumTopTwoPlayerScores),
		DisplayOrder: friend.DisplayOrder,
		PlayerScores: playerScores,
	}
}

func newPlayerScores(players []db.Player, playerNameScores map[db.ID]nameScore) []PlayerScore {
	playerScores := make([]PlayerScore, len(players))
	for i, player := range players {
		playerScores[i] = newPlayerScore(player, playerNameScores[player.ID])
	}
	displayOrder := func(i int) int { return playerScores[i].DisplayOrder }
	sort.Slice(playerScores, func(i, j int) bool {
		return displayOrder(i) < displayOrder(j)
	})
	return playerScores
}

func newPlayerScore(player db.Player, playerNameScore nameScore) PlayerScore {
	return PlayerScore{
		ID:           player.ID,
		Name:         playerNameScore.name,
		Score:        playerNameScore.score,
		DisplayOrder: player.DisplayOrder,
		SourceID:     player.SourceID,
	}
}

func getFriendScore(playerScores []PlayerScore, onlySumTopTwoPlayerScores bool) int {
	scores := make([]int, len(playerScores))
	for i, playerNameScore := range playerScores {
		scores[i] = playerNameScore.Score
	}
	if onlySumTopTwoPlayerScores && len(scores) > 2 {
		sort.Ints(scores) // ex: 1 2 3 4 5
		scores = scores[len(scores)-2:]
	}
	friendScore := 0
	for _, score := range scores {
		friendScore += score
	}
	return friendScore
}

func playerNameScoresFromFieldMaps(players []db.Player, names map[db.SourceID]string, stats map[db.SourceID]int) map[db.ID]nameScore {
	playerNameScores := make(map[db.ID]nameScore, len(players))
	for _, player := range players {
		playerNameScores[player.ID] = nameScore{
			name:  names[player.SourceID],
			score: stats[player.SourceID],
		}
	}
	return playerNameScores
}

func playerNameScoresFromSourceIDMap(players []db.Player, sourceIDNameScores map[db.SourceID]nameScore) map[db.ID]nameScore {
	playerNameScores := make(map[db.ID]nameScore, len(players))
	for _, player := range players {
		playerNameScores[player.ID] = sourceIDNameScores[player.SourceID]
	}
	return playerNameScores
}
