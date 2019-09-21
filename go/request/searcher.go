package request

import "github.com/jacobpatterson1549/nate-mlb/go/db"

type (
	// Searcher requests PlayerSearchResults
	searcher interface {
		search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error)
	}

	// PlayerSearchResult contains information about the result for a searched player.
	PlayerSearchResult struct {
		Name     string
		Details  string
		SourceID db.SourceID
	}
)
