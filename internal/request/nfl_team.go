package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
)

// nflTeamRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflTeamRequestor struct{}

// NflSchedule contains information about NFL teams for a specific year
type NflSchedule struct {
	Teams map[string]NflTeam `json:"nflTeams"`
}

// NflTeam contains information about an NFL team for a specifc year
type NflTeam struct {
	ID     string `json:"nflTeamId"`
	Name   string `json:"fullName"`
	Record string `json:"record"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflTeamRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	nflTeams, err := r.requestNflTeams(fpi.Year)
	if err != nil {
		return scoreCategory, err
	}
	teamNames := make(map[int]string, len(nflTeams))
	teamStats := make(map[int]int, len(nflTeams))
	for id, nflTeam := range nflTeams {
		wins, err := nflTeam.wins()
		if err != nil {
			return scoreCategory, err
		}
		teamNames[id] = nflTeam.Name
		teamStats[id] = wins
	}
	teamNameScores, err := playerNameScores(fpi.Players[pt], teamNames, teamStats)
	if err != nil {
		return scoreCategory, err
	}
	return newScoreCategory(fpi, pt, teamNameScores, false), nil
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
	for id, nflTeam := range nflTeams {
		lowerTeamName := strings.ToLower(nflTeam.Name)
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflTeamSearchResults = append(nflTeamSearchResults, PlayerSearchResult{
				Name:     nflTeam.Name,
				Details:  fmt.Sprintf("%s Record", nflTeam.Record),
				PlayerID: id,
			})
		}
	}
	return nflTeamSearchResults, nil
}

func (r nflTeamRequestor) requestNflTeams(year int) (map[int]NflTeam, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v2/nfl/schedule?season=%d&appKey=test_key_1", year)
	var nflSchedule NflSchedule
	err := request.structPointerFromURL(url, &nflSchedule)
	if err != nil {
		return nil, err
	}
	nflTeams := make(map[int]NflTeam)
	for _, nt := range nflSchedule.Teams {
		id, err := nt.id()
		if err != nil {
			return nil, err
		}
		nflTeams[id] = nt
	}
	return nflTeams, nil
}

func (nt NflTeam) id() (int, error) {
	idI, err := strconv.Atoi(nt.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", nt)
	}
	return idI, nil
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
