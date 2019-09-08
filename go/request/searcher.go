package request

import "github.com/jacobpatterson1549/nate-mlb/go/db"

// Searchers maps PlayerTypes to Searchers for them.
var Searchers = map[db.PlayerType]searcher{
	db.PlayerTypeMlbTeam: mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerSearcher{},
	db.PlayerTypePitcher: mlbPlayerSearcher{},
	db.PlayerTypeNflTeam: nflTeamRequestor{},
	db.PlayerTypeNflQB:   nflPlayerRequestor{},
	db.PlayerTypeNflMisc: nflPlayerRequestor{},
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	SourceID db.SourceID
}

// PlayerSearchResults requests PlayerSearchResults
type searcher interface {
	PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, year int, activePlayersOnly bool) ([]PlayerSearchResult, error)
}
