package request

import "github.com/jacobpatterson1549/nate-mlb/go/db"

type (
	// PlayerSearchResult contains information about the result for a searched player.
	PlayerSearchResult struct {
		Name     string
		Details  string
		SourceID db.SourceID
	}

	// PlayerSearchResults requests PlayerSearchResults
	searcher interface {
		PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, year int, activePlayersOnly bool) ([]PlayerSearchResult, error)
	}
)
