package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
)

// nflTeamRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflTeamRequestor struct{}

// NflTeamsSchedule contains information about NFL teams for a specific year
type NflTeamsSchedule struct {
	Teams map[string]NflTeam `json:"nflTeams"`
}

// NflTeam contains information about an NFL team for a specifc year
type NflTeam struct {
	ID     db.SourceID `json:"nflTeamId,string"`
	Name   string      `json:"fullName"`
	Record string      `json:"record"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflTeamRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	nflTeams, err := r.requestNflTeams(fpi.Year)
	if err != nil {
		return scoreCategory, err
	}
	sourceIDNameScores := make(map[db.SourceID]nameScore, len(nflTeams))
	for sourceID, nflTeam := range nflTeams {
		score, err := nflTeam.wins()
		if err != nil {
			return scoreCategory, err
		}
		sourceIDNameScores[sourceID] = nameScore{
			name:  nflTeam.Name,
			score: score,
		}
	}
	playerNameScores := playerNameScoresFromSourceIDMap(fpi.Players[pt], sourceIDNameScores)
	return newScoreCategory(fpi, pt, playerNameScores, false), nil
}

// PlayerSearchResults implements the Searcher interface
func (r nflTeamRequestor) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	activeYear, err := db.GetActiveYear(pt.SportType())
	if err != nil {
		return nil, err
	}
	nflTeams, err := r.requestNflTeams(activeYear)
	if err != nil {
		return nil, err
	}

	var nflTeamSearchResults []PlayerSearchResult
	lowerQuery := strings.ToLower(playerNamePrefix)
	for sourceID, nflTeam := range nflTeams {
		lowerTeamName := strings.ToLower(nflTeam.Name)
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflTeamSearchResults = append(nflTeamSearchResults, PlayerSearchResult{
				Name:     nflTeam.Name,
				Details:  fmt.Sprintf("%s Record", nflTeam.Record),
				SourceID: sourceID,
			})
		}
	}
	return nflTeamSearchResults, nil
}

func (r nflTeamRequestor) requestNflTeams(year int) (map[db.SourceID]NflTeam, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v2/nfl/schedule?season=%d&appKey=test_key_1", year)
	var nflSchedule NflTeamsSchedule
	err := request.structPointerFromURL(url, &nflSchedule)
	if err != nil {
		return nil, err
	}
	nflTeams := make(map[db.SourceID]NflTeam)
	for _, nflTeam := range nflSchedule.Teams {
		nflTeams[nflTeam.ID] = nflTeam
	}
	return nflTeams, nil
}

func (nt NflTeam) wins() (int, error) {
	recordParts := strings.Split(nt.Record, "-")
	if len(recordParts) == 0 {
		return -1, fmt.Errorf("Could not get Wins for %v", nt)
	}
	winsI, err := strconv.Atoi(recordParts[0])
	if err != nil {
		return -1, fmt.Errorf("Invalid Wins number for %v", nt)
	}
	return winsI, nil
}
