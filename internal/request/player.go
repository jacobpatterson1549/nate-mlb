package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
	"sync"
)

func fillerPlayer() {

}

// PlayerInfoRequest contains invormation about requests for hitter/pitcher names/stats
type PlayerInfoRequest struct {
	playerNames map[int]string
	playerStats map[string]map[int]int
	wg          sync.WaitGroup
	lastError   error
	hasError    bool
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
	Stats []struct {
		Group struct {
			DisplayName string `json:"displayName"`
		} `json:"group"`
		Splits []struct {
			Stat Stat `json:"stat"`
		} `json:"splits"`
	} `json:"stats"`
}

// Stat is used too unmarshal stats for a part of a player stat request
type Stat struct {
	HomeRuns int `json:"homeRuns"`
	Wins     int `json:"wins"`
}

func getPlayerScoreCategory(friends []db.Friend, players []db.Player, playerType db.PlayerType, playerInfoRequest *PlayerInfoRequest) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	playerInfoRequest.wg.Wait()
	if playerInfoRequest.hasError {
		return scoreCategory, playerInfoRequest.lastError
	}
	playerScores, err := playerInfoRequest.getPlayerScores(playerType.Name())
	if err == nil {
		err = scoreCategory.compute(friends, players, playerType, playerScores, true)
	}
	return scoreCategory, err
}

func (pir *PlayerInfoRequest) requestPlayerInfoAsync(players []db.Player, year int) {

	pir.playerNames = make(map[int]string)
	pir.playerStats = make(map[string]map[int]int)
	pir.wg = sync.WaitGroup{}

	// Note that these keys are the same as player_types
	// TODO: make this a private field of player type (DisplayName vs GroupName)
	pir.playerStats["hitting"] = make(map[int]int)
	pir.playerStats["pitching"] = make(map[int]int)

	playerIDsSet := make(map[int]bool)
	playerIDstrings := []string{}
	playerIDInts := []int{}
	for _, player := range players {
		if player.PlayerTypeID == 2 || player.PlayerTypeID == 3 {
			if _, ok := playerIDsSet[player.PlayerID]; !ok {
				playerIDsSet[player.PlayerID] = true
				playerIDstrings = append(playerIDstrings, strconv.Itoa(player.PlayerID))
				playerIDInts = append(playerIDInts, player.PlayerID)
			}
		}
	}

	pir.wg.Add(2)
	go pir.requestPlayerNames(playerIDstrings)
	go pir.requestPlayerStats(playerIDInts, year)
}

func (pir *PlayerInfoRequest) requestPlayerNames(playerIDs []string) {
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIDs, ",")), ",", "%2C")
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

func (pir *PlayerInfoRequest) requestPlayerStats(playerIDs []int, year int) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	wg.Add(len(playerIDs))
	for _, playerID := range playerIDs {
		go func(playerID int, mutex *sync.Mutex) {
			pir.requestPlayerStat(playerID, year, mutex)
			wg.Done()
		}(playerID, &mutex)
	}
	wg.Wait()
	pir.addMissingPlayerStats(playerIDs)
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) requestPlayerStat(playerID int, year int, mutex *sync.Mutex) {
	playerStatsURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people/%d/stats?&season=%d&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins", playerID, year), ",", "%2C")
	playerStats := PlayerStats{}
	err := requestStruct(playerStatsURL, &playerStats)

	if err == nil {
		mutex.Lock()
		err = pir.addPlayerStats(playerID, playerStats)
		mutex.Unlock()
	}

	if err != nil {
		pir.hasError = true
		pir.lastError = err
	}
}

func (pir *PlayerInfoRequest) addPlayerStats(playerID int, playerStats PlayerStats) error {
	for _, stats := range playerStats.Stats {
		for groupDisplayName, groupStatsMap := range pir.playerStats {
			if stats.Group.DisplayName == groupDisplayName {
				splits := stats.Splits
				lastStat := splits[len(splits)-1].Stat
				score, err := lastStat.getScore(groupDisplayName)
				if err != nil {
					return err
				}
				groupStatsMap[playerID] = score
			}
		}
	}
	return nil
}

func (pir *PlayerInfoRequest) addMissingPlayerStats(playerIDs []int) {
	// Some players might not have played for the requested year for the position that was requested.
	// If so, add a 0 as their stat.
	// TODO: This bloats the playerStats map, but it is not a big deal for now.
	for _, playerID := range playerIDs {
		for _, playerStats := range pir.playerStats {
			if _, ok := playerStats[playerID]; !ok {
				playerStats[playerID] = 0
			}
		}
	}
}

func (pir *PlayerInfoRequest) getPlayerScores(groupDisplayName string) (map[int]PlayerScore, error) {
	playerScores := make(map[int]PlayerScore)
	for k, v := range pir.playerStats {
		if k == groupDisplayName {
			for playerID, score := range v {
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
		}
	}
	return playerScores, nil
}

func (s *Stat) getScore(groupDisplayName string) (int, error) {
	// TODO: make seperate requests for pitchers and hitters, and key in on (Stat)function()int
	// (these strings are in the data, so they must be switched on)
	switch groupDisplayName {
	case "hitting":
		return s.HomeRuns, nil
	case "pitching":
		return s.Wins, nil
	default:
		return -1, fmt.Errorf("Unknown stat for groupDisplayName %v", groupDisplayName)
	}
}
