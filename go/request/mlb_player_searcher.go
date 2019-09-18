package request

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	// mlbPlayerSearcher implements the searcher interface
	mlbPlayerSearcher struct {
		requestor requestor
	}

	// MlbPlayerSearch is used to unmarshal a request for information about players by name
	MlbPlayerSearch struct {
		SearchPlayerAll MlbPlayerSearchAll `json:"search_player_all"`
	}

	// MlbPlayerSearchAll is part of a mlbPlayerSearch
	MlbPlayerSearchAll struct {
		QueryResults MlbPlayerSearchQueryResults `json:"queryResults"`
	}

	// MlbPlayerSearchQueryResults  is used to unmarshal a request for information about players by name
	MlbPlayerSearchQueryResults struct {
		TotalSize  int             `json:"totalSize,string"`
		PlayerBios json.RawMessage `json:"row"` // will be []PlayerBio, PlayerBio, or absent
	}

	// MlbPlayerBio contains the results of a player search for a single player
	MlbPlayerBio struct {
		Position     string             `json:"position"`
		BirthCountry string             `json:"birth_country"`
		BirthDate    MlbPlayerBirthDate `json:"birth_date"`
		TeamAbbrev   string             `json:"team_abbrev"`
		PlayerName   string             `json:"name_display_first_last"`
		PlayerID     db.SourceID        `json:"player_id,string"`
	}

	// MlbPlayerBirthDate contains information about a players birthdate including if it is missing
	MlbPlayerBirthDate struct {
		time    time.Time
		missing bool
	}
)

// PlayerSearchResults implements the Searcher interface
func (s *mlbPlayerSearcher) PlayerSearchResults(pt db.PlayerType, playerNamePrefix string, year int, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	activePlayers := "N"
	if activePlayersOnly {
		activePlayers = "Y"
	}
	playerNamePrefix = url.QueryEscape(playerNamePrefix)
	url := strings.ReplaceAll(fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.search_player_all.bam?name_part='%s%%25'&active_sw='%s'&sport_code='mlb'&search_player_all.col_in=player_id&search_player_all.col_in=name_display_first_last&search_player_all.col_in=position&search_player_all.col_in=team_abbrev&search_player_all.col_in=team_abbrev&search_player_all.col_in=birth_country&search_player_all.col_in=birth_date", playerNamePrefix, activePlayers), "'", "%27")
	var mlbPlayerSearchQueryResult MlbPlayerSearch
	err := s.requestor.structPointerFromURL(url, &mlbPlayerSearchQueryResult)
	if err != nil {
		return []PlayerSearchResult{}, err
	}
	return mlbPlayerSearchQueryResult.SearchPlayerAll.QueryResults.getPlayerSearchResults(pt)
}

func (psqr *MlbPlayerSearchQueryResults) getPlayerSearchResults(pt db.PlayerType) ([]PlayerSearchResult, error) {
	var mlbPlayerBios []MlbPlayerBio
	var err error
	switch psqr.TotalSize {
	case 0:
		break
	case 1:
		var playerBio MlbPlayerBio
		err = json.Unmarshal(psqr.PlayerBios, &playerBio)
		mlbPlayerBios = append(mlbPlayerBios, playerBio)
	default:
		err = json.Unmarshal(psqr.PlayerBios, &mlbPlayerBios)
	}

	var playerSearchResults []PlayerSearchResult
	if err != nil {
		return playerSearchResults, err
	}
	for _, mlbPlayerBio := range mlbPlayerBios {
		if mlbPlayerBio.matches(pt) {
			playerSearchResults = append(playerSearchResults, mlbPlayerBio.toPlayerSearchResult())
		}
	}
	return playerSearchResults, nil
}

func (mlbPlayerBio MlbPlayerBio) matches(pt db.PlayerType) bool {
	switch mlbPlayerBio.Position {
	case "P":
		return pt == db.PlayerTypePitcher
	default:
		return pt == db.PlayerTypeHitter
	}
}

func (mlbPlayerBio MlbPlayerBio) toPlayerSearchResult() PlayerSearchResult {
	return PlayerSearchResult{
		Name:     mlbPlayerBio.PlayerName,
		Details:  fmt.Sprintf("team:%s, position:%s, born:%s,%s", mlbPlayerBio.TeamAbbrev, mlbPlayerBio.Position, mlbPlayerBio.BirthCountry, mlbPlayerBio.BirthDate),
		SourceID: mlbPlayerBio.PlayerID,
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.  The quoted string time time is expected to be empty (nil), or be in RFC 3339 format without a zone.
func (mpbd *MlbPlayerBirthDate) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		mpbd.missing = true
		return nil
	}
	var err error
	mpbd.time, err = time.Parse(`"2006-01-02T15:04:05"`, string(data))
	return err
}

// String returns the calendar date (YYYY-MM-DD) of the MlbPlayerBirthDate
func (mpbd MlbPlayerBirthDate) String() string {
	if mpbd.missing {
		return "?"
	}
	return mpbd.time.Format("2006-01-02")
}
