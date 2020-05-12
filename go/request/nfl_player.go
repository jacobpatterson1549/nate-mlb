package request

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// nflPlayerRequester implements the ScoreCategorizer and Searcher interfaces
	nflPlayerRequester struct {
		requester requester
	}

	// NflPlayerSearch contains NflGames for a query
	NflPlayerSearch struct {
		Games map[string]NflGame `json:"games"`
	}

	// NflGame contains active NflPlayers for a particular query
	NflGame struct {
		Season  int                  `json:"season,string"`
		Players map[string]NflPlayer `json:"players"`
	}

	// NflPlayer contains the player info and possibly stats
	NflPlayer struct {
		ID       db.SourceID                `json:"playerId,string"`
		Name     string                     `json:"name"`
		Position string                     `json:"position"`
		Team     string                     `json:"nflTeamAbbr"`
		Stats    map[string]json.RawMessage `json:"stats"`
	}

	// NflPlayerStats contains the stats totals a NflPlayerStat has accumulated during a particular year
	// The meaning of these stats can be found at
	// https://api.fantasy.nfl.com/v2/game/stats?appKey=test_key_1
	NflPlayerStats struct {
		PassingTD   int `json:"6,string"`
		RushingTD   int `json:"15,string"`
		ReceivingTD int `json:"22,string"`
		ReturnTD    int `json:"28,string"`
	}
)

const (
	nflAppKey = "test_key_1"
)

// RequestScoreCategory implements the ScoreCategorizer interface
func (r *nflPlayerRequester) RequestScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	sourceIDs := make(map[db.SourceID]bool, len(players))
	services := make([]map[string]string, len(players))
	for i, player := range players {
		sourceIDs[player.SourceID] = true
		services[i] = map[string]string{
			"playerDetails": fmt.Sprintf("season=%d&playerId=%d", year, player.SourceID),
		}
	}
	var scoreCategory ScoreCategory
	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return scoreCategory, fmt.Errorf("could not build request url for bulk nfl player stats: %w", err)
	}
	uri := fmt.Sprintf("batchservices?appKey=%s&services=%s", nflAppKey, servicesJSON)
	nflPlayerSearch, err := r.requestNflPlayerSearch(uri)
	if err != nil {
		return scoreCategory, err
	}

	sourceIDNameScores := make(map[db.SourceID]nameScore, len(sourceIDs))
	for id, nflPlayer := range nflPlayerSearch.players() {
		if _, ok := sourceIDs[nflPlayer.ID]; ok {
			stats, err := nflPlayer.stats()
			if err != nil {
				return scoreCategory, fmt.Errorf("could not get season stats for player %v: %w", id, err)
			}
			sourceIDNameScores[nflPlayer.ID] = nameScore{
				name:  nflPlayer.Name,
				score: stats.stat(pt),
			}
		}
	}
	playerNameScores := playerNameScoresFromSourceIDMap(players, sourceIDNameScores)
	return newScoreCategory(pt, ptInfo, friends, players, playerNameScores, true), nil
}

// Search implements the Searcher interface
// searches active players
func (r *nflPlayerRequester) Search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	positionIds := "1,2,3,4" // QB,RB,WR,TE
	uri := fmt.Sprintf("players/autocomplete?appKey=%s&positionIds=%s&query=%s", nflAppKey, positionIds, playerNamePrefix)
	nflPlayerSearch, err := r.requestNflPlayerSearch(uri)
	if err != nil {
		return nil, err
	}

	var nflPlayerSearchResults []PlayerSearchResult
	lowerQuery := strings.ToLower(playerNamePrefix)
	for _, nflPlayer := range nflPlayerSearch.players() {
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

func (r *nflPlayerRequester) requestNflPlayerSearch(uri string) (*NflPlayerSearch, error) {
	uri = "https://api.fantasy.nfl.com/v2/" + uri
	nflPlayerSearch := new(NflPlayerSearch)
	err := r.requester.structPointerFromURI(uri, &nflPlayerSearch)
	return nflPlayerSearch, err
}

func (s NflPlayerSearch) players() map[string]NflPlayer {
	for _, g := range s.Games {
		return g.Players
	}
	return nil
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

// stats has special handling to return the first season's stats
// the actual stats are structured like {"week:{YEAR:{WEEK:{K:V...}...}...},"season":{YEAR:{K:V...}...}
func (nflPlayer NflPlayer) stats() (NflPlayerStats, error) {
	var nflPlayerStats NflPlayerStats
	rawStatsYearsMap, ok := nflPlayer.Stats["season"]
	if !ok {
		return nflPlayerStats, fmt.Errorf("no season stats for player %v", nflPlayer.ID)
	}
	var statsYears map[string]NflPlayerStats
	err := json.Unmarshal(rawStatsYearsMap, &statsYears)
	if err != nil {
		return nflPlayerStats, fmt.Errorf("could not unmarshal season stats by year: %w", err)
	}
	for _, nflplayerStats := range statsYears {
		return nflplayerStats, nil
	}
	return nflPlayerStats, fmt.Errorf("no season stats")
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
