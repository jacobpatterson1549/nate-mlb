package main

func searchPlayers(playerTypeID int, playerNamePrefix string) ([]PlayerSearchResult, error) {
	var psr []PlayerSearchResult
	psr = []PlayerSearchResult{
		PlayerSearchResult{"SEA", ".412", 136},
		PlayerSearchResult{"SFG", ".500", 137},
	}
	return psr, nil
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}
