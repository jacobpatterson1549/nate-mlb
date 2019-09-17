package request

import "github.com/jacobpatterson1549/nate-mlb/go/db"

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
