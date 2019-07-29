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

func getStats(friendPlayerInfo FriendPlayerInfo) ([]ScoreCategory, error) {
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
	switch playerType.name {
	case "team":
		return getTeamScoreScategory(friendPlayerInfo, playerType)
	case "hitter":
		return getPlayerScoreCategory(friendPlayerInfo, playerType, playerInfoRequest)
	case "pitcher":
		return getPlayerScoreCategory(friendPlayerInfo, playerType, playerInfoRequest)
	default:
		return ScoreCategory{}, fmt.Errorf("Unknown playerType: %v", playerType.name)
		// return ScoreCategory{}, nil
	}
}

func getTeamScoreScategory(friendPlayerInfo FriendPlayerInfo, teamPlayerType PlayerType) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	teamsJSON := TeamsJSON{}
	err := requestJSON("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103%2C104&season=2019", &teamsJSON)
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
	switch playerType.name {
	case "hitter":
		playerScores, err := playerInfoRequest.getHitterPlayerScores()
		if err != nil {
			return scoreCategory, err
		}
		return scoreCategory, scoreCategory.compute(friendPlayerInfo, playerType, playerScores, true)
	case "pitcher":
		// TODO: sloppy
		playerScores, err := playerInfoRequest.getPitcherPlayerScores()
		if err != nil {
			return scoreCategory, err
		}
		return scoreCategory, scoreCategory.compute(friendPlayerInfo, playerType, playerScores, true)
	default:
		return scoreCategory, fmt.Errorf("Cannot get player scores for player type %v", playerType.id)
	}
}

func requestJSON(url string, v interface{}) error {
	request, err := http.NewRequest("GET", url, nil)
	if err == nil {
		request.Header.Add("Accept", "application/json")
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		response, err := client.Do(request)
		if err == nil {
			defer response.Body.Close()
			err = json.NewDecoder(response.Body).Decode(&v)
		}
	}
	return err
}

func (t *TeamsJSON) getPlayerScores() map[int]PlayerScore {
	playerScores := make(map[int]PlayerScore)
	for _, record := range t.Records {
		for _, teamRecord := range record.TeamRecords {
			playerScores[teamRecord.Team.ID] = PlayerScore{
				Name:  teamRecord.Team.Name,
				Score: teamRecord.Wins,
			}
		}
	}
	return playerScores
}

func (sc *ScoreCategory) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) error {
	sc.Name = playerType.name
	sc.FriendScores = make([]FriendScore, len(friendPlayerInfo.friends))
	for i, friend := range friendPlayerInfo.friends {
		friendScore, err := friend.compute(friendPlayerInfo, playerType, playerScores, onlySumTopTwoPlayerScores)
		if err == nil {
			sc.FriendScores[i] = friendScore
		} else {
			return err
		}
	}
	return nil
}

func (f *Friend) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) (FriendScore, error) {
	friendScore := FriendScore{}

	friendScore.Name = f.name

	friendScore.PlayerScores = []PlayerScore{}
	for _, player := range friendPlayerInfo.players {
		if f.id == player.friendID && playerType.id == player.playerTypeID {
			if playerScore, ok := playerScores[player.playerID]; ok {
				friendScore.PlayerScores = append(friendScore.PlayerScores, playerScore)
			} else {
				return friendScore, fmt.Errorf("No Player score for id = %v", player.playerID)
			}
		}
	}

	score := 0
	if onlySumTopTwoPlayerScores {
		scores := make([]int, len(friendScore.PlayerScores))
		for i, playerScore := range friendScore.PlayerScores {
			scores[i] = playerScore.Score
		}
		sort.Ints(scores) // ex: 1 2 3 4 5
		if len(scores) >= 1 {
			score += scores[len(scores)-1]
			if len(scores) >= 2 {
				score += scores[len(scores)-2]
			}
		}
	} else {
		for _, playerScore := range friendScore.PlayerScores {
			score += playerScore.Score
		}
	}
	friendScore.Score = score

	return friendScore, nil
}

// TODO: how to specify to request playerTypes 2 & 3
func (pir *PlayerInfoRequest) requestPlayerInfoAsync(friendPlayerInfo FriendPlayerInfo) {

	pir.playerNames = make(map[int]string)
	pir.playerStatsJSONs = make(map[int]PlayerStatsJSON)
	pir.wg = sync.WaitGroup{}

	playerIDsSet := make(map[int]bool)
	playerIDStrings := []string{}
	playerIDInts := []int{}
	for _, player := range friendPlayerInfo.players {
		if player.playerTypeID == 2 || player.playerTypeID == 3 {
			if _, ok := playerIDsSet[player.playerID]; !ok {
				playerIDsSet[player.playerID] = true
				playerIDStrings = append(playerIDStrings, strconv.Itoa(player.playerID))
				playerIDInts = append(playerIDInts, player.playerID)
			}
		}
	}

	pir.wg.Add(2)
	go pir.requestPlayerNames(playerIDStrings)
	go pir.requestPlayerStats(playerIDInts)
}

func (pir *PlayerInfoRequest) requestPlayerNames(playerIds []string) {
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIds, ",")), ",", "%2C")
	playerNamesJSON := PlayerNamesJSON{}
	request, err := http.NewRequest("GET", playerNamesURL, nil)
	if err == nil {
		request.Header.Add("Accept", "application/json")
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		response, err := client.Do(request)
		if err == nil {
			defer response.Body.Close()
			err = json.NewDecoder(response.Body).Decode(&playerNamesJSON)
			if err == nil {
				for _, people := range playerNamesJSON.People {
					pir.playerNames[people.ID] = people.FullName
				}
			}
		}
	}
	if err != nil {
		pir.hasError = true
		pir.lastError = err
	}
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) requestPlayerStats(playerIds []int) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	wg.Add(len(playerIds))
	for _, playerID := range playerIds {
		go func(playerID int, mutex *sync.Mutex) {
			go pir.requestPlayerStat(playerID, mutex)
			wg.Done()
		}(playerID, &mutex)
	}
	wg.Wait()
	pir.wg.Done()
}

func (pir *PlayerInfoRequest) requestPlayerStat(playerID int, mutex *sync.Mutex) {
	playerStatsURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people/%d/stats?&season=2019&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins", playerID), ",", "%2C")
	playerStatsJSON := PlayerStatsJSON{}
	request, err := http.NewRequest("GET", playerStatsURL, nil)
	if err == nil {
		request.Header.Add("Accept", "application/json")
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		response, err := client.Do(request)
		if err == nil {
			defer response.Body.Close()
			err = json.NewDecoder(response.Body).Decode(&playerStatsJSON)
			if err == nil {
				mutex.Lock()
				pir.playerStatsJSONs[playerID] = playerStatsJSON
				mutex.Unlock()
			}
		}
	}
	if err != nil {
		pir.hasError = true
		pir.lastError = err
	}
}

func (pir *PlayerInfoRequest) getHitterPlayerScores() (map[int]PlayerScore, error) {
	playerScores := make(map[int]PlayerScore)
	for playerID, playerStatsJSON := range pir.playerStatsJSONs {
		for _, stats := range playerStatsJSON.Stats {
			if stats.Group.DisplayName == "hitting" {
				if name, ok := pir.playerNames[playerID]; ok {
					splits := stats.Splits
					playerScores[playerID] = PlayerScore{
						Name:  name,
						Score: splits[len(splits)-1].Stat.HomeRuns,
					}
				} else {
					return playerScores, fmt.Errorf("No name for player %v", playerID)
				}
			}
		}
	}
	return playerScores, nil
}

// TODO: use shared logic to getHitterPlayerScores() ("pitching", func(Stat))
func (pir *PlayerInfoRequest) getPitcherPlayerScores() (map[int]PlayerScore, error) {
	playerScores := make(map[int]PlayerScore)
	for playerID, playerStatsJSON := range pir.playerStatsJSONs {
		for _, stats := range playerStatsJSON.Stats {
			if stats.Group.DisplayName == "pitching" {
				if name, ok := pir.playerNames[playerID]; ok {
					splits := stats.Splits
					playerScores[playerID] = PlayerScore{
						Name:  name,
						Score: splits[len(splits)-1].Stat.Wins,
					}
				} else {
					return playerScores, fmt.Errorf("No name for player %v", playerID)
				}
			}
		}
	}
	return playerScores, nil
}

// ScoreCategory  contain the FriendScores for each PlayerType
type ScoreCategory struct {
	Name         string
	FriendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	Name         string
	PlayerScores []PlayerScore
	Score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	Name  string
	Score int
}

// PlayerInfoRequest contains invormation about requests for hitter/pitcher names/stats
type PlayerInfoRequest struct {
	playerNames      map[int]string
	playerStatsJSONs map[int]PlayerStatsJSON
	wg               sync.WaitGroup
	lastError        error
	hasError         bool
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
			Stat struct {
				HomeRuns int `json:"homeRuns"`
				Wins     int `json:"wins"`
			} `json:"stat"`
		} `json:"splits"`
	} `json:"stats"`
}
