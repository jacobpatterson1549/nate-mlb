package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strings"
)

// mlbTeamRequestor implemnts the ScoreCategorizer and Searcher interfaces
type mlbTeamRequestor struct{}

// MlbTeams is used to unmarshal a wins request for all teams
type MlbTeams struct {
	Records []MlbTeamRecords `json:"records"`
}

// MlbTeamRecords contain the records for teams
type MlbTeamRecords struct {
	TeamRecords []MlbTeamRecord `json:"teamRecords"`
}

// MlbTeamRecord contain the records for teams in a division of a league of a sport
type MlbTeamRecord struct {
	Team   MlbTeamRecordName `json:"team"`
	Wins   int               `json:"wins"`
	Losses int               `json:"losses"`
}

// MlbTeamRecordName contains the name and id a MlbTeamRecord is for
type MlbTeamRecordName struct {
	Name string      `json:"name"`
	ID   db.SourceID `json:"id"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r mlbTeamRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	teams, err := r.requestMlbTeams(fpi.Year)
	if err != nil {
		return scoreCategory, err
	}
	sourceIDNameScores := make(map[db.SourceID]nameScore)
	for _, record := range teams.Records {
		for _, teamRecord := range record.TeamRecords {
			sourceIDNameScores[teamRecord.Team.ID] = nameScore{
				name:  teamRecord.Team.Name,
				score: teamRecord.Wins,
			}
		}
	}
	playerNameScores := playerNameScoresFromSourceIDMap(fpi.Players[pt], sourceIDNameScores)
	return newScoreCategory(fpi, pt, playerNameScores, false), nil
}

// PlayerSearchResults implements the Searcher interface
func (r mlbTeamRequestor) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	var teamSearchResults []PlayerSearchResult
	activeYear, err := db.GetActiveYear(pt.SportType())
	if err != nil {
		return teamSearchResults, err
	}
	teams, err := r.requestMlbTeams(activeYear)
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

func (r mlbTeamRequestor) requestMlbTeams(year int) (MlbTeams, error) {
	var mlbTeams MlbTeams
	url := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year), ",", "%2C")
	err := request.structPointerFromURL(url, &mlbTeams)
	return mlbTeams, err
}
