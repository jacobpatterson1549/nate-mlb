package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func searchPlayers(playerTypeID int, playerNamePrefix string) ([]PlayerSearchResult, error) {
	switch {
	case playerTypeID == 1:
		return searchTeams(playerNamePrefix)
	case playerTypeID == 2 || playerTypeID == 3:
		return searchPlayerNames(playerNamePrefix)
	default:
		return []PlayerSearchResult{}, fmt.Errorf("cannot search for playerTypeID %d", playerTypeID)
	}
}

func searchTeams(query string) ([]PlayerSearchResult, error) {
	teamSearchResults := []PlayerSearchResult{}
	activeYear, err := getActiveYear()
	if err != nil {
		return teamSearchResults, err
	}
	teamsJSON, err := requestTeamsJSON(activeYear)
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

func searchPlayerNames(playerNamePrefix string) ([]PlayerSearchResult, error) {
	playerSearchResults := []PlayerSearchResult{}
	currentYear := getUtcTime().Year()
	activeYear, err := getActiveYear()
	if err != nil {
		return playerSearchResults, err
	}
	activePlayers := "Y"
	if currentYear != activeYear {
		activePlayers = "N"
	}
	url := strings.ReplaceAll(fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.search_player_all.bam?name_part='%s%%25'&active_sw='%s'&sport_code='mlb'&search_player_all.col_in=player_id&search_player_all.col_in=name_display_first_last&search_player_all.col_in=position&search_player_all.col_in=team_abbrev&search_player_all.col_in=team_abbrev&search_player_all.col_in=birth_country&search_player_all.col_in=birth_date", playerNamePrefix, activePlayers), "'", "%27")
	response, err := request(url)
	if err != nil {
		return playerSearchResults, err
	}
	defer response.Body.Close()

	// sometimes the rows are '[]Row' and sometimes it is 'Row' (a singular row) -- so this hack must exist.  read the body, try to prase it twice
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return playerSearchResults, err
	}
	psmj := PlayerSearchMultipleJSON{}
	err = json.Unmarshal(b, &psmj)
	if err != nil {
		pssj := PlayerSearchSingleJSON{}
		err = json.Unmarshal(b, &pssj)
		if err != nil {
			return playerSearchResults, err
		}
		psr, err := pssj.SearchPlayerAll.QueryResults.Row.toPlayerSearchResult()
		if err != nil {
			return playerSearchResults, err
		}
		playerSearchResults = append(playerSearchResults, psr)
	} else {
		playerSearchResults = make([]PlayerSearchResult, len(psmj.SearchPlayerAll.QueryResults.Rows))
		for i, row := range psmj.SearchPlayerAll.QueryResults.Rows {
			psr, err := row.toPlayerSearchResult()
			if err != nil {
				return playerSearchResults, err
			}
			playerSearchResults[i] = psr
		}
	}

	return playerSearchResults, nil
}

func (row Row) toPlayerSearchResult() (PlayerSearchResult, error) {
	var psr PlayerSearchResult
	birthDate := row.BirthDate[:10]             // YYYY-MM-DD
	playerID, err := strconv.Atoi(row.PlayerID) // all players must have valid ids, ignore bad ids
	if err == nil {
		psr = PlayerSearchResult{
			Name:     row.PlayerName,
			Details:  fmt.Sprintf("team:%s, position:%s, born:%s,%s", row.TeamAbbrev, row.Position, row.BirthCountry, birthDate),
			PlayerID: playerID,
		}
	}
	return psr, nil
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}

// PlayerSearchMultipleJSON contain the results of a player search that returns more than one row.
type PlayerSearchMultipleJSON struct {
	SearchPlayerAll struct {
		QueryResults struct {
			Rows []Row `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// PlayerSearchSingleJSON contain the results of a player search that returns more than one row.
type PlayerSearchSingleJSON struct {
	SearchPlayerAll struct {
		QueryResults struct {
			Row Row `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// Row contains the results of a player search for a single player
type Row struct {
	Position     string `json:"position"`
	BirthCountry string `json:"birth_country"`
	BirthDate    string `json:"birth_date"`
	TeamAbbrev   string `json:"team_abbrev"`
	PlayerName   string `json:"name_display_first_last"`
	PlayerID     string `json:"player_id"`
}
