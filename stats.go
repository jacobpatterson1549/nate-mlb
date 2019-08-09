package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getStats() ([]ScoreCategory, error) {

	friendPlayerInfo, err := getFriendPlayerInfo()
	if err != nil {
		return nil, err
	}

	numCategories := len(friendPlayerInfo.playerTypes)
	scoreCategories := make([]ScoreCategory, numCategories)
	var wg sync.WaitGroup
	wg.Add(numCategories)
	var lastError error
	playerInfoRequest := PlayerInfoRequest{}
	playerInfoRequest.requestPlayerInfoAsync(friendPlayerInfo)
	for i, playerType := range friendPlayerInfo.playerTypes {
		go func(i int, playerType PlayerType) {
			scoreCategory, err := getScoreCategory(friendPlayerInfo, playerType, &playerInfoRequest)
			if err != nil {
				lastError = err
			} else {
				scoreCategories[i] = scoreCategory
			}
			wg.Done()
		}(i, playerType)
	}
	wg.Wait()
	return scoreCategories, lastError
}

func getScoreCategory(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerInfoRequest *PlayerInfoRequest) (ScoreCategory, error) {
	switch playerType.id {
	case playerTypeTeam:
		return getTeamScoreScategory(friendPlayerInfo, playerType)
	case playerTypeHitting, playerTypePitching:
		return getPlayerScoreCategory(friendPlayerInfo, playerType, playerInfoRequest)
	default:
		return ScoreCategory{}, fmt.Errorf("Unknown playerType: %v", playerType.name)
	}
}

func getTeamScoreScategory(friendPlayerInfo FriendPlayerInfo, teamPlayerType PlayerType) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	teamsJSON, err := requestTeamsJSON(friendPlayerInfo.year)
	if err == nil {
		playerScores := teamsJSON.getPlayerScores()
		err = scoreCategory.compute(friendPlayerInfo, teamPlayerType, playerScores, false)
	}
	return scoreCategory, err
}

func getPlayerScoreCategory(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerInfoRequest *PlayerInfoRequest) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	playerInfoRequest.wg.Wait()
	if playerInfoRequest.hasError {
		return scoreCategory, playerInfoRequest.lastError
	}
	playerScores, err := playerInfoRequest.getPlayerScores(playerType.name)
	if err == nil {
		err = scoreCategory.compute(friendPlayerInfo, playerType, playerScores, true)
	}
	return scoreCategory, err
}

func request(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return client.Do(request)
}

func requestJSON(url string, v interface{}) error {
	response, err := request(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(&v)
}

func requestTeamsJSON(year int) (TeamsJSON, error) {
	teamsJSON := TeamsJSON{}
	url := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year), ",", "%2C")
	return teamsJSON, requestJSON(url, &teamsJSON)
}

func (t *TeamsJSON) getPlayerScores() map[int]PlayerScore {
	playerScores := make(map[int]PlayerScore)
	for _, record := range t.Records {
		for _, teamRecord := range record.TeamRecords {
			playerScores[teamRecord.Team.ID] = PlayerScore{
				PlayerName: teamRecord.Team.Name,
				PlayerID:   teamRecord.Team.ID,
				Score:      teamRecord.Wins,
			}
		}
	}
	return playerScores
}

func (sc *ScoreCategory) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) error {
	sc.Name = playerType.name
	sc.Description = playerType.description
	sc.PlayerTypeID = playerType.id
	sc.FriendScores = make([]FriendScore, len(friendPlayerInfo.friends))
	for i, friend := range friendPlayerInfo.friends {
		friendScore, err := friend.compute(friendPlayerInfo, playerType, playerScores, onlySumTopTwoPlayerScores)
		if err != nil {
			return err
		}
		sc.FriendScores[i] = friendScore
	}
	return nil
}

func (f *Friend) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) (FriendScore, error) {
	friendScore := FriendScore{}

	friendScore.FriendName = f.name
	friendScore.FriendID = f.id

	friendScore.PlayerScores = []PlayerScore{}
	for _, player := range friendPlayerInfo.players {
		if f.id == player.friendID && playerType.id == player.playerTypeID {
			playerScore, ok := playerScores[player.playerID]
			if !ok {
				return friendScore, fmt.Errorf("No Player score for id = %v, type = %v", player.playerID, playerType.name)
			}
			playerScoreWithID := PlayerScore{
				PlayerName: playerScore.PlayerName,
				PlayerID:   playerScore.PlayerID,
				ID:         player.id,
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

func (pir *PlayerInfoRequest) requestPlayerInfoAsync(friendPlayerInfo FriendPlayerInfo) {

	pir.playerNames = make(map[int]string)
	pir.playerStats = make(map[string]map[int]int)
	pir.wg = sync.WaitGroup{}

	// Note that these keys are the same as player_types
	pir.playerStats["hitting"] = make(map[int]int)
	pir.playerStats["pitching"] = make(map[int]int)

	playerIDsSet := make(map[int]bool)
	playerIDstrings := []string{}
	playerIDInts := []int{}
	for _, player := range friendPlayerInfo.players {
		if player.playerTypeID == 2 || player.playerTypeID == 3 {
			if _, ok := playerIDsSet[player.playerID]; !ok {
				playerIDsSet[player.playerID] = true
				playerIDstrings = append(playerIDstrings, strconv.Itoa(player.playerID))
				playerIDInts = append(playerIDInts, player.playerID)
			}
		}
	}

	pir.wg.Add(2)
	go pir.requestPlayerNames(playerIDstrings)
	go pir.requestPlayerStats(playerIDInts, friendPlayerInfo.year)
}

func (pir *PlayerInfoRequest) requestPlayerNames(playerIDs []string) {
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIDs, ",")), ",", "%2C")
	playerNamesJSON := PlayerNamesJSON{}
	err := requestJSON(playerNamesURL, &playerNamesJSON)
	if err == nil {
		pir.addPlayerNames(playerNamesJSON)
	} else {
		pir.hasError = true
		pir.lastError = err
	}
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) addPlayerNames(playerNamesJSON PlayerNamesJSON) {
	for _, people := range playerNamesJSON.People {
		pir.playerNames[people.ID] = people.FullName
	}
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
	playerStatsJSON := PlayerStatsJSON{}
	err := requestJSON(playerStatsURL, &playerStatsJSON)

	if err == nil {
		mutex.Lock()
		err = pir.addPlayerStats(playerID, playerStatsJSON)
		mutex.Unlock()
	}

	if err != nil {
		pir.hasError = true
		pir.lastError = err
	}
}

func (pir *PlayerInfoRequest) addPlayerStats(playerID int, playerStatsJSON PlayerStatsJSON) error {
	for _, stats := range playerStatsJSON.Stats {
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

// PlayerInfoRequest contains invormation about requests for hitter/pitcher names/stats
type PlayerInfoRequest struct {
	playerNames map[int]string
	playerStats map[string]map[int]int
	wg          sync.WaitGroup
	lastError   error
	hasError    bool
}

// TeamsJSON is used to unmarshal a wins request for all teams
type TeamsJSON struct {
	Records []struct {
		TeamRecords []struct {
			Team struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"team"`
			Wins int `json:"wins"`
		} `json:"teamRecords"`
	} `json:"records"`
}

// PlayerNamesJSON is used to unmarshal a request for player names
type PlayerNamesJSON struct {
	People []struct {
		ID       int    `json:"id"`
		FullName string `json:"fullName"`
	} `json:"people"`
}

// PlayerStatsJSON is used to unmarshal a player homeRuns/wins request
type PlayerStatsJSON struct {
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
