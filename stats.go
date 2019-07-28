package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
	teamsJSON, err := requestTeamsJSON()
	if err == nil {
		playerScores := teamsJSON.getPlayerScores()
		err = scoreCategory.compute(friendPlayerInfo, teamPlayerType, playerScores, false)
	}
	return scoreCategory, err
}

// TODO get all player score for a category in bulk
// &player_id=%27605483%27
func getPlayerScoreCategory(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, url string, scoreKey string) (ScoreCategory, error) {
	// TODO
	return ScoreCategory{}, nil
}

func requestTeamsJSON() (TeamsJSON, error) {
	teamsJSON := TeamsJSON{}
	request, err := http.NewRequest("GET", "http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103%2C104&season=2019", nil)
	if err == nil {
		request.Header.Add("Accept", "application/json")
		client := &http.Client{
			Timeout: 1 * time.Second,
		}
		response, err := client.Do(request)
		if err == nil {
			defer response.Body.Close()
			err = json.NewDecoder(response.Body).Decode(&teamsJSON)
		}
	}
	return teamsJSON, err
}

func (t *TeamsJSON) getPlayerScores() map[int]PlayerScore {
	playerScores := make(map[int]PlayerScore)
	for _, record := range t.Records {
		for _, teamRecord := range record.TeamRecords {
			playerScores[teamRecord.Team.ID] = PlayerScore{
				name:  teamRecord.Team.Name,
				score: teamRecord.Wins,
			}
		}
	}
	return playerScores
}

func (sc *ScoreCategory) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) error {
	sc.name = playerType.name
	sc.friendScores = make([]FriendScore, len(friendPlayerInfo.friends))
	for i, friend := range friendPlayerInfo.friends {
		friendScore, err := friend.compute(friendPlayerInfo, playerType, playerScores, onlySumTopTwoPlayerScores)
		if err == nil {
			sc.friendScores[i] = friendScore
		} else {
			return err
		}
	}
	return nil
}

func (f *Friend) compute(friendPlayerInfo FriendPlayerInfo, playerType PlayerType, playerScores map[int]PlayerScore, onlySumTopTwoPlayerScores bool) (FriendScore, error) {
	friendScore := FriendScore{}

	friendScore.name = f.name

	friendScore.playerScores = []PlayerScore{}
	for _, player := range friendPlayerInfo.players {
		if f.id == player.friendID && playerType.id == player.playerTypeID {
			if playerScore, ok := playerScores[player.playerID]; ok {
				friendScore.playerScores = append(friendScore.playerScores, playerScore)
			} else {
				return friendScore, fmt.Errorf("No Player scor for id = %v", player.playerID)
			}
		}
	}

	score := 0
	if onlySumTopTwoPlayerScores {
		scores := make([]int, len(friendScore.playerScores))
		for i, playerScore := range friendScore.playerScores {
			scores[i] = playerScore.score
		}
		sort.Ints(scores) // ex: 1 2 3 4 5
		if len(scores) >= 1 {
			score += scores[len(scores)-1]
			if len(scores) >= 2 {
				score += scores[len(scores)-2]
			}
		}
	} else {
		for _, playerScore := range friendScore.playerScores {
			score += playerScore.score
		}
	}
	friendScore.score = score

	return friendScore, nil
}

// ScoreCategory  contain the FriendScores for each PlayerType
type ScoreCategory struct {
	name         string
	friendScores []FriendScore
}

// FriendScore contain the scores for a Friend for a PlayerType
type FriendScore struct {
	name         string
	playerScores []PlayerScore
	score        int
}

// PlayerScore is the score for a particular Player
type PlayerScore struct {
	name  string
	score int
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
