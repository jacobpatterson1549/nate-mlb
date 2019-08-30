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
				Name string      `json:"name"`
				ID   db.SourceID `json:"id,int"` // TODO: is the `int` necessary?
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
	teamNames := make(map[db.SourceID]string)
	teamStats := make(map[db.SourceID]int)
	for _, record := range teams.Records {
		for _, teamRecord := range record.TeamRecords {
			teamNames[teamRecord.Team.ID] = teamRecord.Team.Name
			teamStats[teamRecord.Team.ID] = teamRecord.Wins
		}
	}
	teamNameScores, err := playerNameScores(fpi.Players[pt], teamNames, teamStats)
	if err != nil {
		return scoreCategory, err
	}
	return newScoreCategory(fpi, pt, teamNameScores, false), nil
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
					SourceID: teamRecord.Team.ID,
				})
			}
		}
	}
	return teamSearchResults, nil
}

func (r mlbTeamRequestor) requestTeams(year int) (Teams, error) {
	var teams Teams // TODO: rename to MlbTeams, rename func
	url := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year), ",", "%2C")
	err := request.structPointerFromURL(url, &teams)
	return teams, err
}
