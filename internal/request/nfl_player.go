package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
)

// nflPlayerRequestor implemnts the ScoreCategorizer and Searcher interfaces
type nflPlayerRequestor struct {
	playerType db.PlayerType
}

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
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r nflPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	var scoreCategory ScoreCategory
	playerScores := make(map[int]*PlayerScore)
	scoreCategory.populate(fpi.Friends, fpi.Players, pt, playerScores, true)
	return scoreCategory, nil
}

// PlayerSearchResults implements the Searcher interface
func (r nflPlayerRequestor) PlayerSearchResults(st db.SportType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	activeYear, err := db.GetActiveYear(st)
	if err != nil {
		return nil, err
	}
	nflPlayerDetails, err := r.requestNflPlayerDetails(activeYear)
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
				Details:  fmt.Sprintf("TODO: details"),
				PlayerID: id,
			})
		}
	}
	return nflPlayerSearchResults, nil
}

func (r nflPlayerRequestor) requestNflPlayerDetails(year int) (map[int]NflPlayerInfo, error) {
	var nflPlayerList NflPlayerList
	maxCount := 10000
	url := fmt.Sprintf("https://api.fantasy.nfl.com/v1/players/researchinfo?format=json&count=%d&season=%d", maxCount, year)
	err := requestStruct(url, &nflPlayerList)
	if err != nil {
		return nil, err
	}
	nflPlayerDetails := make(map[int]NflPlayerInfo)
	for _, npi := range nflPlayerList.Players {
		id, err := npi.id()
		if err != nil {
			return nil, err
		}
		nflPlayerDetails[id] = npi
	}
	return nflPlayerDetails, nil
}

func (npi NflPlayerInfo) id() (int, error) {
	idI, err := strconv.Atoi(npi.ID)
	if err != nil {
		return -1, fmt.Errorf("Invalid Id number for %v", npi)
	}
	return idI, nil
}

func (npi NflPlayerInfo) fullName() string {
	return fmt.Sprintf("%s %s", npi.FirstName, npi.LastName)
}
