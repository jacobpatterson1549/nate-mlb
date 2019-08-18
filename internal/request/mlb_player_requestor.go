package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
	"sync"
)

// mlbPlayerRequestor contains invormation about requests for hitter/pitcher names/stats
type mlbPlayerRequestor struct {
	playerType db.PlayerType
}

// PlayerNames is used to unmarshal a request for player names
type PlayerNames struct {
	People []struct {
		ID       int    `json:"id"`
		FullName string `json:"fullName"`
	} `json:"people"`
}

// PlayerStats is used to unmarshal a player homeRuns/wins request
type PlayerStats struct {
	PlayerTypeStats []PlayerTypeStat `json:"stats"`
}

// PlayerTypeStat contains stats for a type of position for a player
type PlayerTypeStat struct {
	Group struct {
		DisplayName string `json:"displayName"`
	} `json:"group"`
	Splits []struct {
		Stat Stat `json:"stat"`
	} `json:"splits"`
}

// Stat contains a stat for a particular team the player has been on, or is the sum of stats if it is the last one
type Stat struct {
	HomeRuns int `json:"homeRuns"`
	Wins     int `json:"wins"`
}

// RequestScoreCategory implements the requestor interface
func (r mlbPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	playerScores, lastError := r.requestPlayerScores(fpi.Players, fpi.Year)
	var scoreCategory ScoreCategory
	if lastError != nil {
		return scoreCategory, lastError
	}
	err := scoreCategory.populate(fpi.Friends, fpi.Players, r.playerType, playerScores, true)
	return scoreCategory, err
}

func (r mlbPlayerRequestor) requestPlayerScores(players []db.Player, year int) (map[int]PlayerScore, error) {
	playerScores := make(map[int]PlayerScore)
	for _, player := range players {
		if player.PlayerType == r.playerType {
			playerScores[player.PlayerID] = PlayerScore{
				PlayerID: player.PlayerID,
			}
		}
	}

	var wg sync.WaitGroup
	var lastError error
	wg.Add(2)
	go r.requestPlayerNames(playerScores, &lastError, &wg)
	go r.requestPlayerStats(playerScores, year, &lastError, &wg)
	wg.Wait()
	return playerScores, lastError
}

func (r *mlbPlayerRequestor) requestPlayerNames(playerScores map[int]PlayerScore, lastError *error, wg *sync.WaitGroup) {
	playerIDStrings := make([]string, len(playerScores))
	i := 0
	for playerID := range playerScores {
		playerIDStrings[i] = strconv.Itoa(playerID)
		i++
	}
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIDStrings, ",")), ",", "%2C")
	var playerNames PlayerNames
	err := requestStruct(playerNamesURL, &playerNames)

	if err != nil {
		*lastError = err
	} else {
		for _, person := range playerNames.People {
			if playerScore, ok := playerScores[person.ID]; ok {
				playerScore.PlayerName = person.FullName
			}
		}
	}
	wg.Done()
}

func (r *mlbPlayerRequestor) requestPlayerStats(playerScores map[int]PlayerScore, year int, lastError *error, wg *sync.WaitGroup) {
	wg.Add(len(playerScores))
	for playerID := range playerScores {
		go r.getPlayerScore(playerID, playerScores, year, lastError, wg)
	}
	wg.Done()
}

func (r *mlbPlayerRequestor) getPlayerScore(playerID int, playerScores map[int]PlayerScore, year int, lastError *error, wg *sync.WaitGroup) {
	score, err := r.requestPlayerScore(playerID, year)
	if err != nil {
		*lastError = err
	} else {
		if playerScore, ok := playerScores[playerID]; ok {
			playerScore.Score = score
		}
	}
	wg.Done()
}

func (r *mlbPlayerRequestor) requestPlayerScore(playerID int, year int) (int, error) {
	playerStatsURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people/%d/stats?&season=%d&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins", playerID, year), ",", "%2C")
	var playerStats PlayerStats
	err := requestStruct(playerStatsURL, &playerStats)
	if err != nil {
		return -1, err
	}
	return playerStats.getScore(r.playerType)
}

func (ps PlayerStats) getScore(playerType db.PlayerType) (int, error) {
	switch playerType {
	case db.PlayerTypeHitter:
		return ps.lastStatScore("hitting", Stat.getHomeRuns), nil
	case db.PlayerTypePitcher:
		return ps.lastStatScore("pitching", Stat.getWins), nil
	default:
		return -1, fmt.Errorf("Cannot get score of playerType %v for player", playerType)
	}
}

func (ps PlayerStats) lastStatScore(groupDisplayName string, score func(Stat) int) int {
	for _, playerTypeStat := range ps.PlayerTypeStats {
		if groupDisplayName == playerTypeStat.Group.DisplayName {
			splits := playerTypeStat.Splits
			if len(splits) > 0 {
				lastStat := splits[len(splits)-1].Stat
				return score(lastStat)
			}
		}
	}
	return 0
}

func (s Stat) getHomeRuns() int {
	return s.HomeRuns
}

func (s Stat) getWins() int {
	return s.Wins
}
