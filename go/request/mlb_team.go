package request

import (
	"fmt"
	"strings"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// mlbTeamRequestor implements the ScoreCategorizer and Searcher interfaces
	mlbTeamRequestor struct {
		requestor requestor
	}

	// MlbTeams is used to unmarshal a wins request for all teams
	MlbTeams struct {
		Records []MlbTeamRecords `json:"records"`
	}

	// MlbTeamRecords contain the records for teams
	MlbTeamRecords struct {
		TeamRecords []MlbTeamRecord `json:"teamRecords"`
	}

	// MlbTeamRecord contain the records for teams in a division of a league of a sport
	MlbTeamRecord struct {
		Team   MlbTeamRecordName `json:"team"`
		Wins   int               `json:"wins"`
		Losses int               `json:"losses"`
	}

	// MlbTeamRecordName contains the name and id a MlbTeamRecord is for
	MlbTeamRecordName struct {
		Name string      `json:"name"`
		ID   db.SourceID `json:"id"`
	}
)

// RequestScoreCategory implements the ScoreCategorizer interface
func (r *mlbTeamRequestor) RequestScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	teams, err := r.requestMlbTeams(year)
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
	playerNameScores := playerNameScoresFromSourceIDMap(players, sourceIDNameScores)
	return newScoreCategory(pt, ptInfo, friends, players, playerNameScores, false), nil
}

// Search implements the Searcher interface
func (r *mlbTeamRequestor) Search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	var teamSearchResults []PlayerSearchResult
	teams, err := r.requestMlbTeams(year)
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

func (r *mlbTeamRequestor) requestMlbTeams(year int) (MlbTeams, error) {
	var mlbTeams MlbTeams
	uri := strings.ReplaceAll(
		fmt.Sprintf(
			"http://statsapi.mlb.com/api/v1/standings/regularSeason?leagueId=103,104&season=%d", year),
		",",
		"%2C")
	err := r.requestor.structPointerFromURI(uri, &mlbTeams)
	return mlbTeams, err
}
