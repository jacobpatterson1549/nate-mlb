package request

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// nflTeamRequester implements the ScoreCategorizer and Searcher interfaces
	nflTeamRequester struct {
		requester requester
	}

	// NflTeamsSchedule contains information about NFL teams for a specific year
	NflTeamsSchedule struct {
		Teams map[db.SourceID]NflTeam `json:"nflTeams"`
	}

	// NflTeam contains information about an NFL team for a specifc year
	NflTeam struct {
		Name   string `json:"fullName"`
		Record string `json:"record"`
	}
)

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflTeamRequester) RequestScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	nflTeams, err := r.requestNflTeams(year)
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
	playerNameScores := playerNameScoresFromSourceIDMap(players, sourceIDNameScores)
	return newScoreCategory(pt, ptInfo, friends, players, playerNameScores, false), nil
}

// Search implements the Searcher interface
func (r nflTeamRequester) Search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	nflTeams, err := r.requestNflTeams(year)
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
	sourceID := func(i int) int { return int(nflTeamSearchResults[i].SourceID) }
	sort.Slice(nflTeamSearchResults, func(i, j int) bool {
		return sourceID(i) < sourceID(j)
	})
	return nflTeamSearchResults, nil
}

func (r *nflTeamRequester) requestNflTeams(year int) (map[db.SourceID]NflTeam, error) {
	uri := fmt.Sprintf("https://api.fantasy.nfl.com/v2/nfl/schedule?appKey=%s&season=%d", nflAppKey, year)
	var nflSchedule NflTeamsSchedule
	err := r.requester.structPointerFromURI(uri, &nflSchedule)
	if err != nil {
		return nil, err
	}
	nflTeams := make(map[db.SourceID]NflTeam)
	for nflTeamID, nflTeam := range nflSchedule.Teams {
		nflTeams[nflTeamID] = nflTeam
	}
	return nflTeams, nil
}

func (nflTeam NflTeam) wins() (int, error) {
	recordParts := strings.Split(nflTeam.Record, "-")
	winsI, err := strconv.Atoi(recordParts[0])
	if err != nil {
		return -1, fmt.Errorf("invalid Wins number for %v", nflTeam)
	}
	return winsI, nil
}
