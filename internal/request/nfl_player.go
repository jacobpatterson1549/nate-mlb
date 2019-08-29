package request

import (
	"fmt"
	"log"
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

	playerNames := make(chan playerName, len(playerScores))
	playerStats := make(chan playerStat, len(playerScores))
	quit := make(chan error)
	go r.requestPlayerNames(pt, fpi.Year, playerScores, playerNames, quit)
	go r.requestPlayerStats(pt, fpi.Year, playerScores, playerStats, quit)

	i := 0
	var scoreCategory ScoreCategory
	for {
		select {
		case err := <-quit:
			return scoreCategory, err
		case playerName := <-playerNames:
			if playerScore, ok := playerScores[playerName.id]; ok {
				playerScore.PlayerName = playerName.name
				i++
			}
		case playerStat := <-playerStats:
			if playerScore, ok := playerScores[playerStat.id]; ok {
				playerScore.Score = playerStat.stat
				i++
			}
		}
		if i == len(playerScores)*2 {
			scoreCategory.populate(fpi.Friends, fpi.Players, pt, playerScores, true)
			return scoreCategory, nil
		}
	}
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
	for id, nflPlayerDetail := range nflPlayerDetails { // TODO: rename nflPlayerDetail
		lowerTeamName := strings.ToLower(nflPlayerDetail.fullName())
		if strings.Contains(lowerTeamName, lowerQuery) {
			nflPlayerSearchResults = append(nflPlayerSearchResults, PlayerSearchResult{
				Name:     nflPlayerDetail.fullName(),
				Details:  fmt.Sprintf("Team: %s, Position: %s", nflPlayerDetail.Team, nflPlayerDetail.Position),
				PlayerID: id,
			})
		}
	}
	return nflPlayerSearchResults, nil
}

func (r *nflPlayerRequestor) requestNflPlayerDetails(pt db.PlayerType, year int) (map[int]NflPlayerDetail, error) {
	var nflPlayerList NflPlayerList
	maxCount := 10000
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/researchinfo?format=json&count=%d&season=%d", maxCount, year)
	err := request.structPointerFromURL(url, &nflPlayerList)
	if err != nil {
		return nil, err
	}
	nflPlayerDetails := make(map[int]NflPlayerDetail)
	for _, nflPlayerDetail := range nflPlayerList.Players {
		if nflPlayerDetail.matches(pt) {
			id, err := nflPlayerDetail.id()
			if err != nil {
				return nil, err
			}
			nflPlayerDetails[id] = nflPlayerDetail
		}
	}
	return nflPlayerDetails, nil
}

func (r *nflPlayerRequestor) requestNflPlayerStats(year int) (map[int]NflPlayerStat, error) {
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/stats?statType=seasonStats&season=%d&format=json", year)
	var nflPlayerStatList NflPlayerStatList
	err := request.structPointerFromURL(url, &nflPlayerStatList)
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

func (r *nflPlayerRequestor) requestPlayerNames(pt db.PlayerType, year int, playerIDs map[int]*PlayerScore, playerNames chan<- playerName, quit chan<- error) {
	nflPlayerDetails, err := r.requestNflPlayerDetails(pt, year)
	if err != nil {
		quit <- err
	} else {
		for id := range playerIDs {
			if npd, ok := nflPlayerDetails[id]; ok {
				playerNames <- playerName{
					id:   id,
					name: npd.fullName(),
				}
			} else {
				playerNames <- playerName{id: id}
				log.Println("No player name found for nfl player", id)
			}
		}
	}
}

func (r *nflPlayerRequestor) requestPlayerStats(pt db.PlayerType, year int, playerIDs map[int]*PlayerScore, playerStats chan<- playerStat, quit chan<- error) {
	nflPlayerStats, err := r.requestNflPlayerStats(year)
	if err != nil {
		quit <- err
		return
	}
	var stat int
	for id := range playerIDs {
		if nflPlayerStat, ok := nflPlayerStats[id]; ok {
			stat, err = nflPlayerStat.Stat.stat(pt)
			if err != nil {
				quit <- err
				return
			}
			playerStats <- playerStat{
				id:   id,
				stat: stat,
			}
		} else {
			playerStats <- playerStat{id: id}
			log.Println("No player stat found for nfl player", id)
		}
	}
}

func (nflPlayerDetail *NflPlayerDetail) id() (int, error) {
	idI, err := strconv.Atoi(nflPlayerDetail.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", nflPlayerDetail)
	}
	return idI, nil
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

func (nps *NflPlayerStat) id() (int, error) {
	idI, err := strconv.Atoi(nps.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", nps)
	}
	return idI, nil
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
