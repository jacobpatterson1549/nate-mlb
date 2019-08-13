package request

import (
	"encoding/json"
	"fmt"
	"nate-mlb/internal/db"
	"sort"
	"strings"
	"sync"
	"time"
)

// EtlStats containS ScoreCategories that were stored at a specific time
type EtlStats struct {
	EtlTime         time.Time
	EtlRefreshTime  time.Time
	ScoreCategories []ScoreCategory
}

// ScoreCategory contain the FriendScores for each PlayerType
type ScoreCategory struct {
	Name         string
	Description  string
	PlayerTypeID int
	FriendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	FriendName   string
	FriendID     int
	PlayerScores []PlayerScore
	Score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	PlayerName string
	PlayerID   int
	ID         int
	Score      int
}

// GetEtlStats retrieves, calculates, and caches the player stats
func GetEtlStats() (EtlStats, error) {
	var es EtlStats

	var year int
	etlJSON, err := db.GetEtlStatsJSON()
	if err != nil {
		return es, err
	}
	fetchStats := true
	currentTime := db.GetUtcTime()
	if len(etlJSON) > 0 {
		err = json.Unmarshal([]byte(etlJSON), &es)
		if err != nil {
			return es, fmt.Errorf("problem converting stats from json for year %v: %v", year, err)
		}
		es.EtlRefreshTime = previousMidnight(currentTime)
		fetchStats = es.EtlTime.Before(es.EtlRefreshTime)
	}
	if fetchStats {
		scoreCategories, err := getScoreCategories()
		if err != nil {
			return es, err
		}
		es.ScoreCategories = scoreCategories
		es.EtlTime = currentTime
		etlJSON, err := json.Marshal(es)
		if err != nil {
			return es, fmt.Errorf("problem converting stats to json for year %v: %v", year, err)
		}
		err = db.SetEtlStats(string(etlJSON))
	}
	return es, err
}

func getScoreCategories() ([]ScoreCategory, error) {

	friends, err := db.GetFriends()
	if err != nil {
		return nil, err
	}
	playerTypes, err := db.LoadPlayerTypes()
	if err != nil {
		return nil, err
	}
	players, err := db.GetPlayers()
	if err != nil {
		return nil, err
	}
	activeYear, err := db.GetActiveYear()
	if err != nil {
		return nil, err
	}

	numCategories := len(playerTypes)
	scoreCategories := make([]ScoreCategory, numCategories)
	var wg sync.WaitGroup
	wg.Add(numCategories)
	var lastError error
	playerInfoRequest := PlayerInfoRequest{}
	playerInfoRequest.requestPlayerInfoAsync(players, activeYear)
	for i, playerType := range playerTypes {
		go func(i int, playerType db.PlayerType) {
			var scoreCategory ScoreCategory
			switch playerType {
			case db.PlayerTypeTeam:
				scoreCategory, err = createTeamScoreScategory(friends, players, playerType, activeYear)
			case db.PlayerTypeHitter, db.PlayerTypePitcher:
				scoreCategory, err = createPlayerScoreCategory(friends, players, playerType, &playerInfoRequest)
			default:
				err = fmt.Errorf("unknown playerType: %v", playerType)
			}
			if err != nil {
				lastError = err // ingore earlier errors - the last one set is shown
			} else {
				scoreCategories[i] = scoreCategory
			}
			wg.Done()
		}(i, playerType)
	}
	wg.Wait()
	return scoreCategories, lastError
}

func (sc *ScoreCategory) populate(friends []db.Friend, players []db.Player, playerType db.PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) error {
	sc.Name = playerType.Name()
	sc.Description = playerType.Description()
	sc.PlayerTypeID = int(playerType)
	sc.FriendScores = make([]FriendScore, len(friends))
	for i, friend := range friends {
		friendScore, err := computeFriendScore(friend, players, playerType, playerScores, onlySumTopTwoPlayerScores)
		if err != nil {
			return err
		}
		sc.FriendScores[i] = friendScore
	}
	return nil
}

func computeFriendScore(friend db.Friend, players []db.Player, playerType db.PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) (FriendScore, error) {
	friendScore := FriendScore{}

	friendScore.FriendName = friend.Name
	friendScore.FriendID = friend.ID

	friendScore.PlayerScores = []PlayerScore{}
	for _, player := range players {
		if friend.ID == player.FriendID && playerType == player.PlayerType {
			playerScore, ok := playerScores[player.PlayerID]
			if !ok {
				return friendScore, fmt.Errorf("no Player score for id = %v, type = %v", player.PlayerID, playerType.Name())
			}
			playerScoreWithID := PlayerScore{
				PlayerName: playerScore.PlayerName,
				PlayerID:   playerScore.PlayerID,
				ID:         player.ID,
				Score:      playerScore.Score,
			}
			friendScore.PlayerScores = append(friendScore.PlayerScores, playerScoreWithID)
		}
	}

	scores := make([]int, len(friendScore.PlayerScores))
	for i, playerScore := range friendScore.PlayerScores {
		scores[i] = playerScore.Score
	}
	if onlySumTopTwoPlayerScores && len(scores) > 2 {
		sort.Ints(scores) // ex: 1 2 3 4 5
		scores = scores[len(scores)-2:]
	}
	friendScore.Score = 0
	for _, score := range scores {
		friendScore.Score += score
	}

	return friendScore, nil
}

// previousMidnight returns the previous midnight relative to Honolulu time (UTC-10)
func previousMidnight(t time.Time) time.Time {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, t.Location())
	if midnight.After(t) {
		midnight = midnight.Add(-24 * time.Hour)
	}
	return midnight
}

// GetName implements the server.Tab interface for ScoreCategory
func (sc ScoreCategory) GetName() string {
	return sc.Name
}

// GetID implements the server.Tab interface for ScoreCategory
func (sc ScoreCategory) GetID() string {
	return strings.ReplaceAll(sc.GetName(), " ", "-")
}
