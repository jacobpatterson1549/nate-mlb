package request

import (
	"fmt"
	"log"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
	"sync"
)

// nflPlayerRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflPlayerRequestor struct{}

// NflPlayerList contains information about all the players for a particular year
type NflPlayerList struct {
	Date    string          `json:"lastUpdated"`
	Players []NflPlayerInfo `json:"players"`
}

// NflPlayerInfo is information about a player for a year
type NflPlayerInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Team      string `json:"teamAbbr"`
	Position  string `json:"position"`
}

// NflPlayerStatList contains information about the stats for all players for a particular year
type NflPlayerStatList struct {
	Players []NflPlayerStat `json:"players"`
}

// NflPlayerStat contains the Stats for a nfl player for a particular year
type NflPlayerStat struct {
	ID   string  `json:"id"`
	Stat NflStat `json:"stats"`
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
	playerScores := make(map[int]*PlayerScore)
	for _, player := range fpi.Players {
		if player.PlayerType == pt {
			playerScores[player.PlayerID] = &PlayerScore{
				PlayerID: player.PlayerID,
			}
		}
	}

	var wg sync.WaitGroup
	var lastError error
	wg.Add(2)
	go r.requestPlayerNames(playerScores, pt, fpi.Year, &lastError, &wg)
	go r.requestPlayerStats(playerScores, fpi.Year, pt, &lastError, &wg)
	wg.Wait()

	var scoreCategory ScoreCategory
	if lastError != nil {
		return scoreCategory, lastError
	}
	scoreCategory.populate(fpi.Friends, fpi.Players, pt, playerScores, true)
	return scoreCategory, nil
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
	for id, npi := range nflPlayerDetails {
		lowerTeamName := strings.ToLower(npi.fullName())
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflPlayerSearchResults = append(nflPlayerSearchResults, PlayerSearchResult{
				Name:     npi.fullName(),
				Details:  fmt.Sprintf("Team: %s, Position: %s", npi.Team, npi.Position),
				PlayerID: id,
			})
		}
	}
	return nflPlayerSearchResults, nil
}

func (r *nflPlayerRequestor) requestNflPlayerDetails(pt db.PlayerType, year int) (map[int]NflPlayerInfo, error) {
	var nflPlayerList NflPlayerList
	maxCount := 10000
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/researchinfo?format=json&count=%d&season=%d", maxCount, year)
	err := requestStruct(url, &nflPlayerList)
	if err != nil {
		return nil, err
	}
	nflPlayerDetails := make(map[int]NflPlayerInfo)
	for _, npi := range nflPlayerList.Players {
		if npi.matches(pt) {
			id, err := npi.id()
			if err != nil {
				return nil, err
			}
			nflPlayerDetails[id] = npi
		}
	}
	return nflPlayerDetails, nil
}

func (r *nflPlayerRequestor) requestNflPlayerStats(year int) (map[int]NflPlayerStat, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/stats?statType=seasonStats&season=%d&format=json", year)
	var nflPlayerStatList NflPlayerStatList
	err := requestStruct(url, &nflPlayerStatList)
	if err != nil {
		return nil, err
	}
	nflPlayerStats := make(map[int]NflPlayerStat)
	for _, nps := range nflPlayerStatList.Players {
		id, err := nps.id()
		if err != nil {
			return nil, err
		}
		nflPlayerStats[id] = nps
	}
	return nflPlayerStats, nil
}

func (r *nflPlayerRequestor) requestPlayerNames(playerScores map[int]*PlayerScore, pt db.PlayerType, year int, lastError *error, wg *sync.WaitGroup) {
	nflPlayerDetails, err := r.requestNflPlayerDetails(pt, year)
	if err != nil {
		*lastError = err
	} else {
		for playerID, playerScore := range playerScores {
			nflPlayerInfo, ok := nflPlayerDetails[playerID]
			if ok {
				playerScore.PlayerName = nflPlayerInfo.fullName()
			} else {
				log.Println("No player details (name) found for player", playerID)
			}
		}
	}
	wg.Done()
}

func (r *nflPlayerRequestor) requestPlayerStats(playerScores map[int]*PlayerScore, year int, pt db.PlayerType, lastError *error, wg *sync.WaitGroup) {
	nflPlayerStats, err := r.requestNflPlayerStats(year)
	if err != nil {
		*lastError = err
	} else {
		var score int
		for playerID, playerScore := range playerScores {
			nflPlayerStat, ok := nflPlayerStats[playerID]
			if ok {
				score, err = nflPlayerStat.Stat.score(pt)
				if err != nil {
					log.Println(playerID, err)
				} else {
					playerScore.Score = score
				}
			} else {
				log.Println("No player details (name) found for player", playerID)
			}
		}
	}
	wg.Done()
}

func (npi *NflPlayerInfo) id() (int, error) {
	idI, err := strconv.Atoi(npi.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", npi)
	}
	return idI, nil
}

func (npi *NflPlayerInfo) fullName() string {
	return fmt.Sprintf("%s %s", npi.FirstName, npi.LastName)
}

func (npi *NflPlayerInfo) matches(pt db.PlayerType) bool {
	switch pt {
	case db.PlayerTypeNflQB:
		return npi.Position == "QB"
	case db.PlayerTypeNflRBWR:
		return npi.Position == "RB" || npi.Position == "WR"
	default:
		return false
	}
}

func (nps *NflPlayerStat) id() (int, error) {
	idI, err := strconv.Atoi(nps.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", nps)
	}
	return idI, nil
}

func (ns *NflStat) score(pt db.PlayerType) (int, error) {
	score := 0
	if pt == db.PlayerTypeNflQB && len(ns.PassingTD) != 0 {
		td, err := strconv.Atoi(ns.PassingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get PassingTD from %v", ns)
		}
		score += td
	}
	if (pt == db.PlayerTypeNflQB || pt == db.PlayerTypeNflRBWR) && len(ns.RushingTD) != 0 {
		td, err := strconv.Atoi(ns.RushingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get RushingTD from %v", ns)
		}
		score += td
	}
	if pt == db.PlayerTypeNflRBWR && len(ns.ReceivingTD) != 0 {
		td, err := strconv.Atoi(ns.ReceivingTD)
		if err != nil {
			return score, fmt.Errorf("problem: could not get ReceivingTD from %v", ns)
		}
		score += td
	}
	return score, nil
}
