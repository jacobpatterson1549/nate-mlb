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
	PlayerType   db.PlayerType // Used as an int on the website
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
	SourceID     db.SourceID
}

type playerName struct {
	sourceID db.SourceID
	name     string
}

type playerStat struct {
	sourceID db.SourceID
	stat     int
}

type sourceID int // TODO: DELETEME (use db.SourceID)

type nameScore struct {
	name  string
	score int
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

func newScoreCategory(fpi FriendPlayerInfo, playerType db.PlayerType, playerNameScores map[int]nameScore, onlySumTopTwoPlayerScores bool) ScoreCategory {
	return ScoreCategory{
		Name:         playerType.Name(),
		PlayerType:   playerType,
		Description:  playerType.Description(),
		FriendScores: newFriendScores(fpi, playerType, playerNameScores, onlySumTopTwoPlayerScores),
	}
}

func newFriendScores(fpi FriendPlayerInfo, playerType db.PlayerType, playerNameScores map[int]nameScore, onlySumTopTwoPlayerScores bool) []FriendScore {
	friendPlayers := make(map[int][]db.Player)
	for _, player := range fpi.Players[playerType] {
		friendPlayers[player.FriendID] = append(friendPlayers[player.FriendID], player)
	}
	friendScores := make([]FriendScore, len(fpi.Friends))
	for i, friend := range fpi.Friends {
		friendScores[i] = newFriendScore(friend, friendPlayers[friend.ID], playerNameScores, onlySumTopTwoPlayerScores)
	}
	return friendScores
}

func newFriendScore(friend db.Friend, players []db.Player, playerNameScores map[int]nameScore, onlySumTopTwoPlayerScores bool) FriendScore {
	playerScores := newPlayerScores(players, playerNameScores)
	return FriendScore{
		ID:           friend.ID,
		Name:         friend.Name,
		Score:        getFriendScore(playerScores, onlySumTopTwoPlayerScores),
		DisplayOrder: friend.DisplayOrder,
		PlayerScores: playerScores,
	}
}

func newPlayerScores(players []db.Player, playerNameScores map[int]nameScore) []PlayerScore {
	playerScores := make([]PlayerScore, len(players))
	for i, player := range players {
		playerScores[i] = newPlayerScore(player, playerNameScores[player.ID])
	}
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

func playerNameScores(players []db.Player, names map[db.SourceID]string, stats map[db.SourceID]int) (map[int]nameScore, error) {
	playerNameScores := make(map[int]nameScore, len(players))
	for _, player := range players {
		name, ok := names[player.SourceID]
		if !ok {
			return playerNameScores, fmt.Errorf("No name for player %d", player.ID)
		}
		stat, ok := stats[player.SourceID]
		if !ok {
			return playerNameScores, fmt.Errorf("No stat for player %d", player.ID)
		}
		playerNameScores[player.ID] = nameScore{
			name:  name,
			score: stat,
		}
	}
	return playerNameScores, nil
}
