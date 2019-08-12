package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
	"sync"
)

// PlayerInfoRequest contains invormation about requests for hitter/pitcher names/stats
type PlayerInfoRequest struct {
	playerNames map[int]string
	playerStats map[db.PlayerType]map[int]int
	wg          sync.WaitGroup
	lastError   error
	hasError    bool // TODO: DELETEME
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

func createPlayerScoreCategory(friends []db.Friend, players []db.Player, playerType db.PlayerType, playerInfoRequest *PlayerInfoRequest) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	playerInfoRequest.wg.Wait()
	if playerInfoRequest.hasError {
		return scoreCategory, playerInfoRequest.lastError
	}
	playerScores, err := playerInfoRequest.createPlayerScores(playerType)
	if err == nil {
		err = scoreCategory.compute(friends, players, playerType, playerScores, true)
	}
	return scoreCategory, err
}

func (pir *PlayerInfoRequest) requestPlayerInfoAsync(players []db.Player, year int) {

	pir.playerNames = make(map[int]string)
	pir.playerStats = make(map[db.PlayerType]map[int]int)
	pir.wg = sync.WaitGroup{}

	// Note that these keys are the same as player_types
	pir.playerStats[db.Hitter] = make(map[int]int)
	pir.playerStats[db.Pitcher] = make(map[int]int)

	playerIDs := make(map[int]string)
	for _, player := range players {
		// TODO: make player.PlayerTypeID be a PlayerType and rename to player.playerType
		if player.PlayerTypeID == int(db.Hitter) || player.PlayerTypeID == int(db.Pitcher) {
			if _, ok := playerIDs[player.PlayerID]; !ok {
				playerIDs[player.PlayerID] = strconv.Itoa(player.PlayerID)
			}
			pir.playerStats[db.PlayerType(player.PlayerTypeID)][player.PlayerID] = 0
		}
	}

	pir.wg.Add(2)
	go pir.requestPlayerNames(playerIDs)
	go pir.requestPlayerStats(year)
}

func (pir *PlayerInfoRequest) requestPlayerNames(playerIDs map[int]string) {
	playerIDStrings := make([]string, len(playerIDs))
	i := 0
	for _, playerIDString := range playerIDs {
		playerIDStrings[i] = playerIDString
		i++
	}
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIDStrings, ",")), ",", "%2C")
	playerNames := PlayerNames{}
	err := requestStruct(playerNamesURL, &playerNames)
	if err == nil {
		for _, people := range playerNames.People {
			pir.playerNames[people.ID] = people.FullName
		}
	} else {
		pir.hasError = true
		pir.lastError = err
	}
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) requestPlayerStats(year int) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for playerType, players := range pir.playerStats {
		for playerID := range players {
			go func(playerID int, mutex *sync.Mutex) {
				pir.requestPlayerStat(playerType, playerID, year, mutex)
				wg.Done()
			}(playerID, &mutex)
		}
		wg.Add(len(players))
	}
	wg.Wait()
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) requestPlayerStat(playerType db.PlayerType, playerID int, year int, mutex *sync.Mutex) {
	playerStatsURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people/%d/stats?&season=%d&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins", playerID, year), ",", "%2C")
	playerStats := PlayerStats{}
	err := requestStruct(playerStatsURL, &playerStats)

	if err == nil {
		var score int
		score, err = playerStats.getScore(playerType)
		if err == nil {
			mutex.Lock()
			pir.playerStats[playerType][playerID] = score
			mutex.Unlock()
		}
	}

	if err != nil {
		pir.hasError = true
		pir.lastError = err
	}
}

func (pir *PlayerInfoRequest) createPlayerScores(playerType db.PlayerType) (map[int]PlayerScore, error) {
	playerScores := make(map[int]PlayerScore)
	for playerID, score := range pir.playerStats[playerType] {
		name, ok := pir.playerNames[playerID]
		if !ok {
			return playerScores, fmt.Errorf("No player name for player %v", playerID)
		}
		playerScores[playerID] = PlayerScore{
			PlayerName: name,
			PlayerID:   playerID,
			Score:      score,
		}
	}
	return playerScores, nil
}

func (ps *PlayerStats) getScore(playerType db.PlayerType) (int, error) {
	switch playerType {
	case db.Hitter:
		return ps.lastStat("hitting", func(s *Stat) int { return s.HomeRuns }), nil
	case db.Pitcher:
		return ps.lastStat("pitching", func(s *Stat) int { return s.Wins }), nil
	default:
		return -1, fmt.Errorf("Cannot get score of playerType %v for player", playerType)
	}
}

func (ps *PlayerStats) lastStat(groupDisplayName string, score func(*Stat) int) int {
	for _, playerTypeStat := range ps.PlayerTypeStats {
		if groupDisplayName == playerTypeStat.Group.DisplayName {
			splits := playerTypeStat.Splits
			if len(splits) > 0 {
				lastStat := splits[len(splits)-1].Stat
				return score(&lastStat)
			}
		}
	}
	return 0 // example: In 2019, Luis Severino is a pitcher, but has not played (TODO: Write test for this)
}
