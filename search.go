package main

import (
	"fmt"
	"strings"
)

func searchPlayers(playerTypeID int, playerNamePrefix string) ([]PlayerSearchResult, error) {
	switch playerTypeID {
	case 1:
		return searchTeams(playerNamePrefix)
	default:
		return []PlayerSearchResult{}, fmt.Errorf("cannot search for playerTypeID %d", playerTypeID)
	}
}

func searchTeams(query string) ([]PlayerSearchResult, error) {
	teamSearchResults := []PlayerSearchResult{}
	teamsJSON, err := requestTeamsJSON()
	if err != nil {
		return teamSearchResults, err
	}

	lowerQuery := strings.ToLower(query)
	for _, record := range teamsJSON.Records {
		for _, teamRecord := range record.TeamRecords {
			lowerTeamName := strings.ToLower(teamRecord.Team.Name)
			if strings.Contains(lowerTeamName, lowerQuery) {
				teamSearchResults = append(teamSearchResults, PlayerSearchResult{
					Name:     teamRecord.Team.Name,
					Details:  fmt.Sprintf("%d wins", teamRecord.Wins),
					PlayerID: teamRecord.Team.ID,
				})
			}
		}
	}
	return teamSearchResults, nil
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}
