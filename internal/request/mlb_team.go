package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strings"
)

// mlbTeamRequestor implemnts the ScoreCategorizer and Searcher interfaces
type mlbTeamRequestor struct{}

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

// RequestScoreCategory implements the ScoreCategorizer interface
func (r mlbTeamRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	teams, err := r.requestTeams(fpi.Year)
	if err != nil {
		return scoreCategory, err
	}

	playerScores := make(map[int]*PlayerScore)
	for _, record := range teams.Records {
		for _, teamRecord := range record.TeamRecords {
			playerScores[teamRecord.Team.ID] = &PlayerScore{
				PlayerName: teamRecord.Team.Name,
				PlayerID:   teamRecord.Team.ID,
				Score:      teamRecord.Wins,
			}
		}
	}
	err = scoreCategory.populate(fpi.Friends, fpi.Players, pt, playerScores, false)
	return scoreCategory, err
}

// PlayerSearchResults implements the Searcher interface
func (r mlbTeamRequestor) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	var teamSearchResults []PlayerSearchResult
	activeYear, err := db.GetActiveYear(pt.SportType())
	if err != nil {
		return teamSearchResults, err
	}
	teams, err := r.requestTeams(activeYear)
	if err != nil {
		return teamSearchResults, err
	}

	lowerQuery := strings.ToLower(playerNamePrefix)
	for _, record := range teams.Records {
		for _, teamRecord := range record.TeamRecords {
			lowerTeamName := strings.ToLower(teamRecord.Team.Name)
			if strings.Contains(lowerTeamName, lowerQuery) {
				teamSearchResults = append(teamSearchResults, PlayerSearchResult{
					Name:     teamRecord.Team.Name,
					Details:  fmt.Sprintf("%d - %d Record", teamRecord.Wins, teamRecord.Losses),
					PlayerID: teamRecord.Team.ID,
				})
			}
		}
	}
	return teamSearchResults, nil
}

func (r mlbTeamRequestor) requestTeams(year int) (Teams, error) {
	var teams Teams
	url := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year), ",", "%2C")
	return teams, requestStruct(url, &teams)
}
