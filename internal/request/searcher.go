package request

import "nate-mlb/internal/db"

// Searchers maps PlayerTypes to Searchers for them.
var Searchers = map[db.PlayerType]searcher{
	db.PlayerTypeMlbTeam: mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerSearcher{},
	db.PlayerTypePitcher: mlbPlayerSearcher{},
	db.PlayerTypeNflTeam: nflTeamRequestor{},
	db.PlayerTypeNflQB:   nflPlayerRequestor{},
	db.PlayerTypeNflRBWR: nflPlayerRequestor{},
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}

// PlayerSearchResults requests PlayerSearchResults
type searcher interface {
	PlayerSearchResults(st db.SportType, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error)
}
