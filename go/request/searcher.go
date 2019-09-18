package request

import "github.com/jacobpatterson1549/nate-mlb/go/db"

type (
	// PlayerSearchResults requests PlayerSearchResults
	searcher interface {
		PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, year int, activePlayersOnly bool) ([]PlayerSearchResult, error)
	}

	// PlayerSearchResult contains information about the result for a searched player.
	PlayerSearchResult struct {
		Name     string
		Details  string
		SourceID db.SourceID
	}
)
