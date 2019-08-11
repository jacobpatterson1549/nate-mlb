package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"nate-mlb/internal/db"
	"net/url"
	"strconv"
	"strings"
)

// SearchPlayers finds the PlayerSearchResults for the specified name prefix
func SearchPlayers(playerTypeID int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	switch {
	case playerTypeID == 1:
		return searchTeams(playerNamePrefix)
	case playerTypeID == 2, playerTypeID == 3:
		return searchPlayerNames(playerNamePrefix, activePlayersOnly)
	default:
		return []PlayerSearchResult{}, fmt.Errorf("cannot search for playerTypeID %d", playerTypeID)
	}
}

func searchTeams(query string) ([]PlayerSearchResult, error) {
	teamSearchResults := []PlayerSearchResult{}
	activeYear, err := db.GetActiveYear()
	if err != nil {
		return teamSearchResults, err
	}
	teamsJSON, err := requestTeams(activeYear)
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
					Details:  fmt.Sprintf("%d - %d Record", teamRecord.Wins, teamRecord.Losses),
					PlayerID: teamRecord.Team.ID,
				})
			}
		}
	}
	return teamSearchResults, nil
}

func searchPlayerNames(playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	playerSearchResults := []PlayerSearchResult{}
	activePlayers := "N"
	if activePlayersOnly {
		activePlayers = "Y"
	}
	playerNamePrefix = url.QueryEscape(playerNamePrefix)
	url := strings.ReplaceAll(fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.search_player_all.bam?name_part='%s%%25'&active_sw='%s'&sport_code='mlb'&search_player_all.col_in=player_id&search_player_all.col_in=name_display_first_last&search_player_all.col_in=position&search_player_all.col_in=team_abbrev&search_player_all.col_in=team_abbrev&search_player_all.col_in=birth_country&search_player_all.col_in=birth_date", playerNamePrefix, activePlayers), "'", "%27")
	response, err := request(url)
	if err != nil {
		return playerSearchResults, err
	}
	defer response.Body.Close()

	// sometimes the rows are '[]Row' and sometimes it is 'Row' (a singular row) -- so this hack must exist.  read the body, try to prase it twice
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return playerSearchResults, fmt.Errorf("problem reading response body from reqeust to %v: %v", url, err)
	}
	psmj := MultiplePlayerSearchResult{}
	err = json.Unmarshal(b, &psmj)
	if err != nil {
		// ignore the error
		pssj := SinglePlayerSearchResult{}
		err = json.Unmarshal(b, &pssj)
		if err != nil {
			return playerSearchResults, fmt.Errorf("problem reading json when requesting %v: %v", url, err)
		}
		psr, err := pssj.SearchPlayerAll.QueryResults.PlayerBio.toPlayerSearchResult()
		if err != nil {
			return playerSearchResults, err
		}
		playerSearchResults = append(playerSearchResults, psr)
	} else {
		playerSearchResults = make([]PlayerSearchResult, len(psmj.SearchPlayerAll.QueryResults.PlayerBios))
		for i, row := range psmj.SearchPlayerAll.QueryResults.PlayerBios {
			psr, err := row.toPlayerSearchResult()
			if err != nil {
				return playerSearchResults, err
			}
			playerSearchResults[i] = psr
		}
	}

	return playerSearchResults, nil
}

func (row PlayerBio) toPlayerSearchResult() (PlayerSearchResult, error) {
	var psr PlayerSearchResult
	birthDate := row.BirthDate[:10]             // YYYY-MM-DD
	playerID, err := strconv.Atoi(row.PlayerID) // all players must have valid ids, ignore bad ids
	if err != nil {
		return psr, fmt.Errorf("problem converting playerId (%v) to number for playerSearch %v: %v", playerID, row, err)
	}

	psr = PlayerSearchResult{
		Name:     row.PlayerName,
		Details:  fmt.Sprintf("team:%s, position:%s, born:%s,%s", row.TeamAbbrev, row.Position, row.BirthCountry, birthDate),
		PlayerID: playerID,
	}
	return psr, nil
}

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}

// MultiplePlayerSearchResult contain the results of a player search that returns more than one row.
type MultiplePlayerSearchResult struct {
	SearchPlayerAll struct {
		QueryResults struct {
			PlayerBios []PlayerBio `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// SinglePlayerSearchResult contain the results of a player search that returns more than one row.
type SinglePlayerSearchResult struct {
	SearchPlayerAll struct {
		QueryResults struct {
			PlayerBio PlayerBio `json:"row"`
		} `json:"queryResults"`
	} `json:"search_player_all"`
}

// PlayerBio contains the results of a player search for a single player
type PlayerBio struct {
	Position     string `json:"position"`
	BirthCountry string `json:"birth_country"`
	BirthDate    string `json:"birth_date"`
	TeamAbbrev   string `json:"team_abbrev"`
	PlayerName   string `json:"name_display_first_last"`
	PlayerID     string `json:"player_id"`
}
