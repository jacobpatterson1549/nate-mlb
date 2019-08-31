package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strings"
)

// nflPlayerRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflPlayerRequestor struct{}

// NflPlayerList contains information about the stats for all players for a particular year
type NflPlayerList struct {
	Players []NflPlayer `json:"players"`
}

// NflPlayer contains the Names Stats for a nfl player for a particular year
type NflPlayer struct {
	ID       db.SourceID    `json:"id,string"`
	Name     string         `json:"name"`
	Position string         `json:"position"`
	Team     string         `json:"teamAbbr"`
	Stats    NflPlayerStats `json:"stats"`
}

// NflPlayerStats contains the stats totals a NflPlayerStat has accumulated during a particular year
// The meaning of these stats can be found at
// https://api.fantasy.nfl.com/v1/game/stats?format=json
type NflPlayerStats struct {
	PassingTD   int `json:"6,string"`
	RushingTD   int `json:"15,string"`
	ReceivingTD int `json:"22,string"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	sourceIDs := make(map[db.SourceID]bool, len(fpi.Players[pt]))
	for _, player := range fpi.Players[pt] {
		sourceIDs[player.SourceID] = true
	}
	var scoreCategory ScoreCategory
	nflPlayerList, err := r.requestNflPlayerList(fpi.Year)
	if err != nil {
		return scoreCategory, err
	}

	sourceIDNameScores := make(map[db.SourceID]nameScore, len(sourceIDs))
	for _, nflPlayer := range nflPlayerList.Players {
		if _, ok := sourceIDs[nflPlayer.ID]; ok {
			score, err := nflPlayer.Stats.stat(pt)
			if err != nil {
				return scoreCategory, err
			}
			sourceIDNameScores[nflPlayer.ID] = nameScore{
				name:  nflPlayer.Name,
				score: score,
			}
		}
	}
	playerNameScores := playerNameScoresFromSourceIDMap(fpi.Players[pt], sourceIDNameScores)
	return newScoreCategory(fpi, pt, playerNameScores, false), nil
}

// PlayerSearchResults implements the Searcher interface
func (r nflPlayerRequestor) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	activeYear, err := db.GetActiveYear(pt.SportType())
	if err != nil {
		return nil, err
	}
	nflPlayerList, err := r.requestNflPlayerList(activeYear)
	if err != nil {
		return nil, err
	}

	var nflPlayerSearchResults []PlayerSearchResult
	lowerQuery := strings.ToLower(playerNamePrefix)
	for _, nflPlayer := range nflPlayerList.Players {
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

func (r nflPlayerRequestor) requestNflPlayerList(year int) (*NflPlayerList, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/stats?statType=seasonStats&season=%d&format=json", year)
	nflPlayerList := new(NflPlayerList)
	err := request.structPointerFromURL(url, &nflPlayerList)
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

func (nflPlayerStat NflPlayerStats) stat(pt db.PlayerType) (int, error) {
	score := 0
	if pt == db.PlayerTypeNflQB {
		score += nflPlayerStat.PassingTD
	}
	if pt == db.PlayerTypeNflQB || pt == db.PlayerTypeNflMisc {
		score += nflPlayerStat.RushingTD
	}
	if pt == db.PlayerTypeNflMisc {
		score += nflPlayerStat.ReceivingTD
	}
	return score, nil
}
