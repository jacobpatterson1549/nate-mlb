package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
)

// nflPlayerRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflPlayerRequestor struct{}

// NflPlayerList contains information about all the players for a particular year
type NflPlayerList struct {
	Date    string            `json:"lastUpdated"`
	Players []NflPlayerDetail `json:"players"`
}

// NflPlayerDetail is information about a player for a year
type NflPlayerDetail struct {
	ID        db.SourceID `json:"id,string"`
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	Team      string      `json:"teamAbbr"`
	Position  string      `json:"position"`
}

// NflPlayerStatList contains information about the stats for all players for a particular year
type NflPlayerStatList struct {
	Players []NflPlayerStat `json:"players"`
}

// NflPlayerStat contains the Stats for a nfl player for a particular year
type NflPlayerStat struct {
	ID   db.SourceID `json:"id,string"`
	Stat NflStat     `json:"stats"`
}

// NflStat contains the stats totals a NflPlayerStat has accumulated during a particular year
// The meaning of these stats can be found at
// https://api.fantasy.nfl.com/v1/game/stats?format=json
type NflStat struct {
	PassingTD   string `json:"6"`
	RushingTD   string `json:"15"`
	ReceivingTD string `json:"22"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	sourceIDs := make(map[db.SourceID]bool)
	for _, player := range fpi.Players[pt] {
		sourceIDs[player.SourceID] = true
	}
	numPlayers := len(sourceIDs)
	playerNames := make(map[db.SourceID]string, numPlayers)
	playerStats := make(map[db.SourceID]int, numPlayers)
	playerNamesCh := make(chan playerName, numPlayers)
	playerStatsCh := make(chan playerStat, numPlayers)
	quit := make(chan error)
	go r.requestPlayerNames(pt, fpi.Year, sourceIDs, playerNamesCh, quit)
	go r.requestPlayerStats(pt, fpi.Year, sourceIDs, playerStatsCh, quit)

	i := 0
	var scoreCategory ScoreCategory
	for {
		select {
		case err := <-quit:
			return scoreCategory, err
		case playerName := <-playerNamesCh:
			playerNames[playerName.sourceID] = playerName.name
		case playerStat := <-playerStatsCh:
			playerStats[playerStat.sourceID] = playerStat.stat
		}
		i++
		if i == numPlayers*2 {
			break
		}
	}
	mlbPlayerNameScores, err := playerNameScores(fpi.Players[pt], playerNames, playerStats)
	if err != nil {
		return scoreCategory, err
	}
	return newScoreCategory(fpi, pt, mlbPlayerNameScores, true), nil
}

// PlayerSearchResults implements the Searcher interface
func (r nflPlayerRequestor) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	activeYear, err := db.GetActiveYear(pt.SportType())
	if err != nil {
		return nil, err
	}
	nflPlayerDetails, err := r.requestNflPlayerDetails(pt, activeYear)
	if err != nil {
		return nil, err
	}

	var nflPlayerSearchResults []PlayerSearchResult
	lowerQuery := strings.ToLower(playerNamePrefix)
	for sourceID, nflPlayerDetail := range nflPlayerDetails {
		lowerTeamName := strings.ToLower(nflPlayerDetail.fullName())
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflPlayerSearchResults = append(nflPlayerSearchResults, PlayerSearchResult{
				Name:     nflPlayerDetail.fullName(),
				Details:  fmt.Sprintf("Team: %s, Position: %s", nflPlayerDetail.Team, nflPlayerDetail.Position),
				SourceID: sourceID,
			})
		}
	}
	return nflPlayerSearchResults, nil
}

func (r *nflPlayerRequestor) requestNflPlayerDetails(pt db.PlayerType, year int) (map[db.SourceID]NflPlayerDetail, error) {
	var nflPlayerList NflPlayerList
	maxCount := 10000
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/researchinfo?format=json&count=%d&season=%d", maxCount, year)
	err := request.structPointerFromURL(url, &nflPlayerList)
	if err != nil {
		return nil, err
	}
	nflPlayerDetails := make(map[db.SourceID]NflPlayerDetail)
	for _, nflPlayerDetail := range nflPlayerList.Players {
		if nflPlayerDetail.matches(pt) {
			nflPlayerDetails[nflPlayerDetail.ID] = nflPlayerDetail
		}
	}
	return nflPlayerDetails, nil
}

func (r *nflPlayerRequestor) requestNflPlayerStats(year int) (map[db.SourceID]NflPlayerStat, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/stats?statType=seasonStats&season=%d&format=json", year)
	var nflPlayerStatList NflPlayerStatList
	err := request.structPointerFromURL(url, &nflPlayerStatList)
	if err != nil {
		return nil, err
	}
	nflPlayerStats := make(map[db.SourceID]NflPlayerStat)
	for _, nflPlayerStat := range nflPlayerStatList.Players {
		nflPlayerStats[nflPlayerStat.ID] = nflPlayerStat
	}
	return nflPlayerStats, nil
}

func (r *nflPlayerRequestor) requestPlayerNames(pt db.PlayerType, year int, sourceIDs map[db.SourceID]bool, playerNames chan<- playerName, quit chan<- error) {
	nflPlayerDetails, err := r.requestNflPlayerDetails(pt, year)
	if err != nil {
		quit <- err
		return
	}
	for sourceID := range sourceIDs {
		var fullName string
		if npd, ok := nflPlayerDetails[sourceID]; ok {
			fullName = npd.fullName()
		}
		playerNames <- playerName{
			sourceID: sourceID,
			name:     fullName,
		}
	}
}

func (r *nflPlayerRequestor) requestPlayerStats(pt db.PlayerType, year int, sourceIDs map[db.SourceID]bool, playerStats chan<- playerStat, quit chan<- error) {
	nflPlayerStats, err := r.requestNflPlayerStats(year)
	if err != nil {
		quit <- err
		return
	}
	for sourceID := range sourceIDs {
		var stat int
		if nflPlayerStat, ok := nflPlayerStats[sourceID]; ok {
			stat, err = nflPlayerStat.Stat.stat(pt)
			if err != nil {
				quit <- err
				return
			}
		}
		playerStats <- playerStat{
			sourceID: sourceID,
			stat:     stat,
		}
	}
}

func (nflPlayerDetail *NflPlayerDetail) fullName() string {
	return fmt.Sprintf("%s %s", nflPlayerDetail.FirstName, nflPlayerDetail.LastName)
}

func (nflPlayerDetail *NflPlayerDetail) matches(pt db.PlayerType) bool {
	switch pt {
	case db.PlayerTypeNflQB:
		return nflPlayerDetail.Position == "QB"
	case db.PlayerTypeNflMisc:
		return nflPlayerDetail.Position == "RB" || nflPlayerDetail.Position == "WR" || nflPlayerDetail.Position == "TE"
	default:
		return false
	}
}

func (ns *NflStat) stat(pt db.PlayerType) (int, error) {
	score := 0
	if pt == db.PlayerTypeNflQB && len(ns.PassingTD) != 0 {
		td, err := strconv.Atoi(ns.PassingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get PassingTD from %v", ns)
		}
		score += td
	}
	if (pt == db.PlayerTypeNflQB || pt == db.PlayerTypeNflMisc) && len(ns.RushingTD) != 0 {
		td, err := strconv.Atoi(ns.RushingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get RushingTD from %v", ns)
		}
		score += td
	}
	if pt == db.PlayerTypeNflMisc && len(ns.ReceivingTD) != 0 {
		td, err := strconv.Atoi(ns.ReceivingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get ReceivingTD from %v", ns)
		}
		score += td
	}
	return score, nil
}
