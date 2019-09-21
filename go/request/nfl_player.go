package request

import (
	"fmt"
	"strings"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// nflPlayerRequestor implements the ScoreCategorizer and Searcher interfaces
	nflPlayerRequestor struct {
		requestor requestor
	}

	// NflPlayerList contains information about the stats for all players for a particular year
	NflPlayerList struct {
		Players []NflPlayer `json:"players"`
	}

	// NflPlayer contains the Names Stats for a nfl player for a particular year
	NflPlayer struct {
		ID       db.SourceID    `json:"id,string"`
		Name     string         `json:"name"`
		Position string         `json:"position"`
		Team     string         `json:"teamAbbr"`
		Stats    NflPlayerStats `json:"stats"`
	}

	// NflPlayerStats contains the stats totals a NflPlayerStat has accumulated during a particular year
	// The meaning of these stats can be found at
	// https://api.fantasy.nfl.com/v1/game/stats?format=json
	NflPlayerStats struct {
		PassingTD   int `json:"6,string"`
		RushingTD   int `json:"15,string"`
		ReceivingTD int `json:"22,string"`
		ReturnTD    int `json:"28,string"`
	}
)

// RequestScoreCategory implements the ScoreCategorizer interface
func (r *nflPlayerRequestor) requestScoreCategory(pt db.PlayerType, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	sourceIDs := make(map[db.SourceID]bool, len(players))
	for _, player := range players {
		sourceIDs[player.SourceID] = true
	}
	var scoreCategory ScoreCategory
	nflPlayerList, err := r.requestNflPlayerList(year)
	if err != nil {
		return scoreCategory, err
	}

	sourceIDNameScores := make(map[db.SourceID]nameScore, len(sourceIDs))
	for _, nflPlayer := range nflPlayerList.Players {
		if _, ok := sourceIDs[nflPlayer.ID]; ok {
			sourceIDNameScores[nflPlayer.ID] = nameScore{
				name:  nflPlayer.Name,
				score: nflPlayer.Stats.stat(pt),
			}
		}
	}
	playerNameScores := playerNameScoresFromSourceIDMap(players, sourceIDNameScores)
	return newScoreCategory(pt, friends, players, playerNameScores, true), nil
}

// PlayerSearchResults implements the Searcher interface
func (r *nflPlayerRequestor) playerSearchResults(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	nflPlayerList, err := r.requestNflPlayerList(year)
	if err != nil {
		return nil, err
	}

	var nflPlayerSearchResults []PlayerSearchResult
	lowerQuery := strings.ToLower(playerNamePrefix)
	for _, nflPlayer := range nflPlayerList.Players {
		if !nflPlayer.matches(pt) {
			continue
		}
		lowerTeamName := strings.ToLower(nflPlayer.Name)
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflPlayerSearchResults = append(nflPlayerSearchResults, PlayerSearchResult{
				Name:     nflPlayer.Name,
				Details:  fmt.Sprintf("Team: %s, Position: %s", nflPlayer.Team, nflPlayer.Position),
				SourceID: nflPlayer.ID,
			})
		}
	}
	return nflPlayerSearchResults, nil
}

func (r *nflPlayerRequestor) requestNflPlayerList(year int) (*NflPlayerList, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/stats?statType=seasonStats&season=%d&format=json", year)
	nflPlayerList := new(NflPlayerList)
	err := r.requestor.structPointerFromURL(url, &nflPlayerList)
	return nflPlayerList, err
}

func (nflPlayer NflPlayer) matches(pt db.PlayerType) bool {
	switch nflPlayer.Position {
	case "QB":
		return pt == db.PlayerTypeNflQB
	case "RB", "WR", "TE":
		return pt == db.PlayerTypeNflMisc
	default:
		return false
	}
}

func (nflPlayerStat NflPlayerStats) stat(pt db.PlayerType) int {
	score := 0
	if pt == db.PlayerTypeNflQB {
		score += nflPlayerStat.PassingTD
	}
	if pt == db.PlayerTypeNflQB || pt == db.PlayerTypeNflMisc {
		score += nflPlayerStat.RushingTD
	}
	if pt == db.PlayerTypeNflMisc {
		score += nflPlayerStat.ReceivingTD
		score += nflPlayerStat.ReturnTD
	}
	return score
}
