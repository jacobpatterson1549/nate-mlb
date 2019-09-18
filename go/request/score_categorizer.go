package request

import (
	"fmt"
	"sort"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// ScoreCategorizer requests data for and creates a ScoreCategory for the FriendPlayerInfo
	ScoreCategorizer interface {
		RequestScoreCategory(pt db.PlayerType, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error)
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

func newScoreCategory(pt db.PlayerType, friends []db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) ScoreCategory {
	return ScoreCategory{
		Name:         pt.Name(),
		PlayerType:   pt,
		Description:  pt.Description(),
		FriendScores: newFriendScores(pt, friends, players, playerNameScores, onlySumTopTwoPlayerScores),
	}
}

func newFriendScores(pt db.PlayerType, friends []db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) []FriendScore {
	friendPlayers := make(map[db.ID][]db.Player, len(players))
	for _, player := range players {
		friendPlayers[player.FriendID] = append(friendPlayers[player.FriendID], player)
	}
	friendScores := make([]FriendScore, len(friends))
	for i, friend := range friends {
		friendScores[i] = newFriendScore(pt, friend, friendPlayers[friend.ID], playerNameScores, onlySumTopTwoPlayerScores)
	}
	sort.Slice(friendScores, func(i, j int) bool {
		return friendScores[i].DisplayOrder < friendScores[j].DisplayOrder
	})
	return friendScores
}

func newFriendScore(pt db.PlayerType, friend db.Friend, players []db.Player, playerNameScores map[db.ID]nameScore, onlySumTopTwoPlayerScores bool) FriendScore {
	playerScores := newPlayerScores(players, playerNameScores)
	return FriendScore{
		ID:           friend.ID,
		Name:         friend.Name,
		ScoreType:    pt.ScoreType(),
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
	sort.Slice(playerScores, func(i, j int) bool {
		return playerScores[i].DisplayOrder < playerScores[j].DisplayOrder
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

func playerNameScoresFromFieldMaps(players []db.Player, names map[db.SourceID]string, stats map[db.SourceID]int) (map[db.ID]nameScore, error) {
	playerNameScores := make(map[db.ID]nameScore, len(players))
	for _, player := range players {
		name, ok := names[player.SourceID]
		if !ok {
			return playerNameScores, fmt.Errorf("no name for player %d", player.ID)
		}
		stat, ok := stats[player.SourceID]
		if !ok {
			return playerNameScores, fmt.Errorf("no stat for player %d", player.ID)
		}
		playerNameScores[player.ID] = nameScore{
			name:  name,
			score: stat,
		}
	}
	return playerNameScores, nil
}

func playerNameScoresFromSourceIDMap(players []db.Player, sourceIDNameScores map[db.SourceID]nameScore) map[db.ID]nameScore {
	playerNameScores := make(map[db.ID]nameScore, len(players))
	for _, player := range players {
		playerNameScores[player.ID] = sourceIDNameScores[player.SourceID]
	}
	return playerNameScores
}
