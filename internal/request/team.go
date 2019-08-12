package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strings"
)

// Teams is used to unmarshal a wins request for all teams
type Teams struct {
	Records []struct {
		TeamRecords []struct {
			Team struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"team"`
			Wins   int `json:"wins"`
			Losses int `json:"losses"`
		} `json:"teamRecords"`
	} `json:"records"`
}

func requestTeams(year int) (Teams, error) {
	teams := Teams{}
	url := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year), ",", "%2C")
	return teams, requestStruct(url, &teams)
}

func createTeamScoreScategory(friends []db.Friend, players []db.Player, teamPlayerType db.PlayerType, year int) (ScoreCategory, error) {
	scoreCategory := ScoreCategory{}
	teams, err := requestTeams(year)
	if err == nil {
		playerScores := teams.createPlayerScores()
		err = scoreCategory.compute(friends, players, teamPlayerType, playerScores, false)
	}
	return scoreCategory, err
}

func (t *Teams) createPlayerScores() map[int]PlayerScore {
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
