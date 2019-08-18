package request

import "nate-mlb/internal/db"

// Searchers maps PlayerTypes to Searchers for them.
// TODO: combine maps with requestor?  Make shared interface?
var Searchers = map[db.PlayerType]searcher{
	db.PlayerTypeTeam:    mlbTeamRequestor{},
	db.PlayerTypeHitter:  mlbPlayerSearcher{},
	db.PlayerTypePitcher: mlbPlayerSearcher{},
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}

// PlayerSearchResults requests PlayerSearchResults
type searcher interface {
	PlayerSearchResults(playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error)
}
