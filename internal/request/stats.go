package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"sort"
)

// ScoreCategorizers maps PlayerTypes to ScoreCategorizers for them
var ScoreCategorizers = map[db.PlayerType]ScoreCategorizer{
	db.PlayerTypeMlbTeam: mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerRequestor{playerType: db.PlayerTypeHitter},
	db.PlayerTypePitcher: mlbPlayerRequestor{playerType: db.PlayerTypePitcher},
	db.PlayerTypeNflTeam: nflTeamRequestor{},
	db.PlayerTypeNflQB:   nflPlayerRequestor{},
	db.PlayerTypeNflMisc: nflPlayerRequestor{},
}

// ScoreCategorizer requests data for and creates a ScoreCategory for the FriendPlayerInfo
type ScoreCategorizer interface {
	RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error)
}

// ScoreCategory contain the FriendScores for each PlayerType
type ScoreCategory struct {
	Name         string
	Description  string
	DisplayOrder int
	FriendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	ID           int
	Name         string
	Score        int
	DisplayOrder int
	PlayerScores []PlayerScore
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	ID           int
	Name         string
	Score        int
	DisplayOrder int
	SourceID     sourceID
}

type playerName struct {
	id   int
	name string
}

type playerStat struct {
	id   int
	stat int
}

type sourceID int

// NameScore is a base score structure: it has an id, name, score, and display order
type NameScore struct { // TOOD: make package-private
	ID           int
	Name         string
	Score        int
	DisplayOrder int
}

// FriendPlayerInfo is a helper pojo of information about what is in a ScoreCategory
type FriendPlayerInfo struct {
	Friends []db.Friend
	Players map[db.PlayerType][]db.Player
	Year    int
}

// NewFriendPlayerInfo creates a FriendPlayerInfo from Friends, Players, PlayerTypes, and a year
func NewFriendPlayerInfo(friends []db.Friend, players []db.Player, playerTypes []db.PlayerType, year int) FriendPlayerInfo {
	playersMap := make(map[db.PlayerType][]db.Player, len(playerTypes))
	for _, player := range players {
		playersMap[player.PlayerType] = append(playersMap[player.PlayerType], player)
	}
	return FriendPlayerInfo{
		Friends: friends,
		Players: playersMap,
		Year:    year,
	}
}

func newScoreCategory(fpi FriendPlayerInfo, playerType db.PlayerType, playerNameScores map[int]NameScore, onlySumTopTwoPlayerScores bool) ScoreCategory {
	return ScoreCategory{
		Name:         playerType.Name(),
		DisplayOrder: playerType.DisplayOrder(),
		Description:  playerType.Description(),
		FriendScores: newFriendScores(fpi, playerType, playerNameScores, onlySumTopTwoPlayerScores),
	}
}

func newFriendScores(fpi FriendPlayerInfo, playerType db.PlayerType, playerNameScores map[int]NameScore, onlySumTopTwoPlayerScores bool) []FriendScore {
	friendScores := make([]FriendScore, len(fpi.Friends))
	for i, friend := range fpi.Friends {
		friendPlayers := []db.Player{}
		for _, player := range fpi.Players[playerType] { // TODO: loop is annoying
			if player.FriendID == friend.ID {
				friendPlayers = append(friendPlayers, player)
			}
		}
		friendScores[i] = newFriendScore(friend, friendPlayers, playerNameScores, onlySumTopTwoPlayerScores)
	}
	return friendScores
}

func newFriendScore(friend db.Friend, players []db.Player, playerNameScores map[int]NameScore, onlySumTopTwoPlayerScores bool) FriendScore {
	playerScores := newPlayerScores(players, playerNameScores)
	return FriendScore{
		ID:           friend.ID,
		Name:         friend.Name,
		Score:        getFriendScore(playerScores, onlySumTopTwoPlayerScores),
		DisplayOrder: friend.DisplayOrder,
		PlayerScores: playerScores,
	}
}

func newPlayerScores(players []db.Player, playerNameScores map[int]NameScore) []PlayerScore {
	playerScores := make([]PlayerScore, len(players))
	for i, player := range players {
		playerScores[i] = newPlayerScore(player, playerNameScores[player.ID])
	}
	return playerScores
}

func newPlayerScore(player db.Player, playerNameScore NameScore) PlayerScore {
	return PlayerScore{
		ID:           player.ID,
		Name:         playerNameScore.Name,
		Score:        playerNameScore.Score,
		DisplayOrder: player.DisplayOrder,
		SourceID:     player.PlayerID, // TODO: rename playerID to sourceID EVERYWHERE (ui, ws, db)
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

func playerNameScores(players []db.Player, names map[int]string, stats map[int]int) (map[int]NameScore, error) {
	playerNameScores := make(map[int]NameScore, len(players))
	for _, player := range players {
		name, ok := names[player.PlayerID]
		if !ok {
			return playerNameScores, fmt.Errorf("No player name for player %d", player.ID)
		}
		stat, ok := stats[player.PlayerID]
		if !ok {
			return playerNameScores, fmt.Errorf("No player stat for player %d", player.ID)
		}
		playerNameScores[player.ID] = NameScore{
			ID:           player.ID,
			Name:         name,
			Score:        stat,
			DisplayOrder: player.DisplayOrder,
		}
	}
	return playerNameScores, nil
}
