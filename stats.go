package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func getStats(friendPlayerInfo FriendPlayerInfo) ([]ScoreCategory, error) {
	numCategories := len(friendPlayerInfo.playerTypes)
	scoreCategories := make([]ScoreCategory, numCategories)
	var wg sync.WaitGroup
	wg.Add(numCategories)
	var lastError error
	for i, playerType := range friendPlayerInfo.playerTypes {
		go func(i int, playerType PlayerType) {
			scoreCategory, err := getScoreCategory(friendPlayerInfo, playerType)
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

func getScoreCategory(friendPlayerInfo FriendPlayerInfo, playerType PlayerType) (ScoreCategory, error) {
	switch playerType.name {
	case "team":
		return getTeamScoreScategory(friendPlayerInfo, playerType)
	case "batter":
		return getPlayerScoreCategory(friendPlayerInfo, playerType, "http://lookup-service-prod.mlb.com/json/named.sport_hitting_tm.bam?league_list_id=%27mlb%27&game_type=%27R%27&season=%272019%27&sport_hitting_tm.col_in=hr", "hr")
	case "pitcher":
		return getPlayerScoreCategory(friendPlayerInfo, playerType, "http://lookup-service-prod.mlb.com/json/named.sport_pitching_tm.bam?league_list_id=%27mlb%27&game_type=%27R%27&season=%272019%27&sport_pitching_tm.col_in=w", "w")
	default:
		return ScoreCategory{}, fmt.Errorf("Unknown playerType: %v", playerType.name)
	}
}

func getTeamScoreScategory(friendPlayerInfo FriendPlayerInfo, teamPlayerType PlayerType) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	request, err := http.NewRequest("GET", "http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103%2C104&season=2019", nil)
	if err != nil {
		return scoreCategory, err
	}
	request.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	response, err := client.Do(request)
	if err != nil {
		return scoreCategory, err
	}
	defer response.Body.Close()

	// Parse the response to Json
	teamsJSON := TeamsJSON{}
	err = json.NewDecoder(response.Body).Decode(&teamsJSON)
	if err != nil {
		return scoreCategory, err
	}
	// Create lookup map
	teamWins := make(map[int]PlayerScore)
	for _, record := range teamsJSON.Records {
		for _, teamRecord := range record.TeamRecords {
			teamWins[teamRecord.Team.ID] = PlayerScore{playerName: teamRecord.Team.Name, score: teamRecord.Wins}
		}
	}

	// populate the FriendScores
	friendScores := make([]FriendScore, len(friendPlayerInfo.friends))
	friendScoresByID := make(map[int]FriendScore)
	for i, friend := range friendPlayerInfo.friends {
		friendScores[i] = FriendScore{friendName: friend.name, playerScores: []PlayerScore{}}
		friendScoresByID[friend.id] = friendScores[i]
	}
	for _, player := range friendPlayerInfo.players {
		if player.playerTypeID == teamPlayerType.id {
			playerScore, ok := teamWins[player.playerID]
			if !ok {
				return scoreCategory, fmt.Errorf("No team wins for team id=%v", player.playerID)
			}
			friendScore, ok := friendScoresByID[player.friendID]
			if !ok {
				return scoreCategory, fmt.Errorf("No friend with id=%v", player.friendID)
			}
			// TODO: Not working correctly
			playerScores := friendScore.playerScores
			playerScores = append(playerScores, playerScore)
			friendScore.playerScores = playerScores
		}
	}

	// Caculate the actual scores
	for _, friendScore := range friendScores {
		winsSum := 0
		for _, playerScore := range friendScore.playerScores {
			winsSum += playerScore.score
		}
		friendScore.score = winsSum
	}

	scoreCategory.name = teamPlayerType.name
	scoreCategory.friendScores = friendScores
	return scoreCategory, nil
}

// TODO get all player score for a category in bulk
// &player_id=%27605483%27
func getPlayerScoreCategory(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, url string, scoreKey string) (ScoreCategory, error) {

	return ScoreCategory{}, nil
}

// ScoreCategory  contain the FriendScores for each PlayerType
type ScoreCategory struct {
	name         string
	friendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	friendName   string
	playerScores []PlayerScore
	score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	playerName string
	score      int
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

// HitterJSON is used to unmarshal a pitcher wins request
type HitterJSON struct {
	Root struct {
		QueryResults struct {
			Row struct {
				HomeRuns int `json:"hr"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_hitting_tm"`
}

// PitcherJSON is used to unmarshal a pitcher wins request
type PitcherJSON struct {
	Root struct {
		QueryResults struct {
			Row struct {
				Wins int `json:"w"`
			} `json:"row"`
		} `json:"queryResults"`
	} `json:"sport_pitching_tm"`
}
