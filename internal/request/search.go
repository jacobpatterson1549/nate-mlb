package request

import (
	"encoding/json"
	"fmt"
	"nate-mlb/internal/db"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PlayerSearchResult contains information about the result for a searched player.
type PlayerSearchResult struct {
	Name     string
	Details  string
	PlayerID int
}

// PlayerSearchQueryResult  is used to unmarshal a request for information about players by name
type PlayerSearchQueryResult struct {
	SearchPlayerAll struct {
		QueryResults QueryResults `json:"queryResults"`
	} `json:"search_player_all"`
}

// QueryResults  is used to unmarshal a request for information about players by name
type QueryResults struct {
	TotalSize  string          `json:"totalSize"`
	PlayerBios json.RawMessage `json:"row"` // will be []PlayerBio, PlayerBio, or absent
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
	var teamSearchResults []PlayerSearchResult
	activeYear, err := db.GetActiveYear()
	if err != nil {
		return teamSearchResults, err
	}
	teams, err := requestTeams(activeYear)
	if err != nil {
		return teamSearchResults, err
	}

	lowerQuery := strings.ToLower(query)
	for _, record := range teams.Records {
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
	activePlayers := "N"
	if activePlayersOnly {
		activePlayers = "Y"
	}
	playerNamePrefix = url.QueryEscape(playerNamePrefix)
	url := strings.ReplaceAll(fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.search_player_all.bam?name_part='%s%%25'&active_sw='%s'&sport_code='mlb'&search_player_all.col_in=player_id&search_player_all.col_in=name_display_first_last&search_player_all.col_in=position&search_player_all.col_in=team_abbrev&search_player_all.col_in=team_abbrev&search_player_all.col_in=birth_country&search_player_all.col_in=birth_date", playerNamePrefix, activePlayers), "'", "%27")
	var playerSearchQueryResult PlayerSearchQueryResult
	err := requestStruct(url, &playerSearchQueryResult)

	if err != nil {
		return []PlayerSearchResult{}, err
	}
	return playerSearchQueryResult.SearchPlayerAll.QueryResults.getPlayerSearchResults()
}

func (psqr QueryResults) getPlayerSearchResults() ([]PlayerSearchResult, error) {
	var playerBios []PlayerBio
	var err error
	switch psqr.TotalSize {
	case "0":
		break
	case "1":
		var playerBio PlayerBio
		err = json.Unmarshal(psqr.PlayerBios, &playerBio)
		playerBios = append(playerBios, playerBio)
	default:
		err = json.Unmarshal(psqr.PlayerBios, &playerBios)
	}

	var playerSearchResults []PlayerSearchResult
	if err == nil {
		playerSearchResults = make([]PlayerSearchResult, len(playerBios))
		for i, pb := range playerBios {
			if err == nil {
				playerSearchResults[i], err = pb.toPlayerSearchResult()
			}
		}
	}
	return playerSearchResults, err
}

func (playerBio PlayerBio) toPlayerSearchResult() (PlayerSearchResult, error) {
	var psr PlayerSearchResult
	bdTime, err := time.Parse("2006-01-02T15:04:05", playerBio.BirthDate)
	if err != nil {
		return psr, fmt.Errorf("problem formatting player birthdate (%v) to time: %v", playerBio.BirthDate, err)
	}
	birthDate := bdTime.Format(time.RFC3339)[:10]     // YYYY-MM-DD
	playerID, err := strconv.Atoi(playerBio.PlayerID) // all players must have valid ids, ignore bad ids
	if err != nil {
		return psr, fmt.Errorf("problem converting playerId (%v) to number for playerSearch %v: %v", playerID, playerBio, err)
	}

	psr = PlayerSearchResult{
		Name:     playerBio.PlayerName,
		Details:  fmt.Sprintf("team:%s, position:%s, born:%s,%s", playerBio.TeamAbbrev, playerBio.Position, playerBio.BirthCountry, birthDate),
		PlayerID: playerID,
	}
	return psr, nil
}
