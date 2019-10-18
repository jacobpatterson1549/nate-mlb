package request

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestMlbPlayerSearchResults(t *testing.T) {
	playerSearchResultsTests := []struct {
		pt                db.PlayerType
		playerNamePrefix  string
		activePlayersOnly bool
		playersJSON       string
		wantErr           bool
		want              []PlayerSearchResult
	}{
		{
			pt:                db.PlayerTypeMlbPitcher,
			playerNamePrefix:  "Hader",
			activePlayersOnly: true,
			playersJSON: `{"search_player_all":{"queryResults":{
				"totalSize":"1",
				"row":{
					"position": "P",
					"birth_country": "USA",
					"birth_date": "1994-04-07T00:00:00",
					"team_abbrev": "MIL",
					"name_display_first_last": "Josh Hader",
					"player_id": "623352"}}}}`,
			want: []PlayerSearchResult{{Name: "Josh Hader", Details: "team:MIL, position:P, born:USA,1994-04-07", SourceID: 623352}},
		},
		{
			pt:                db.PlayerTypeMlbHitter,
			playerNamePrefix:  "jose mart",
			activePlayersOnly: false,
			playersJSON: `{"search_player_all":{"queryResults":{
				"totalSize":"2",
				"row":[{
					"position": "P",
					"birth_country": "Dominican Republic",
					"birth_date": "1971-01-04T00:00:00",
					"team_abbrev": "SD",
					"name_display_first_last": "Jose Martinez",
					"player_id": "118372"
					},
					{
					"position": "2B",
					"birth_country": "Cuba",
					"birth_date": "1942-07-26T00:00:00",
					"team_abbrev": "PIT",
					"name_display_first_last": "Jose Martinez",
					"player_id": "118370"
					}]}}}`, // do not include player 500874 - he is inactive in 2019
			want: []PlayerSearchResult{{Name: "Jose Martinez", Details: "team:PIT, position:2B, born:Cuba,1942-07-26", SourceID: 118370}},
		},
		{
			pt:               db.PlayerTypeMlbHitter,
			playerNamePrefix: "felix",
			wantErr:          true, // no json
		},
		{
			pt:                db.PlayerTypeMlbPitcher,
			playerNamePrefix:  "bartholomew", // no results
			activePlayersOnly: true,
			playersJSON:       `{"search_player_all":{"queryResults":{"totalSize":"0"}}}`,
		},
	}
	for i, test := range playerSearchResultsTests {
		jsonFunc := func(uri string) string {
			switch {
			case test.activePlayersOnly && !strings.Contains(uri, "Y"),
				!test.activePlayersOnly && !strings.Contains(uri, ""):
				t.Errorf("expected uri of request to contain flag for activePlayersOnly (%v): %v", test.activePlayersOnly, uri)
			}
			return test.playersJSON
		}
		r := newMockHTTPRequester(jsonFunc)
		mlbPlayerSearcher := mlbPlayerSearcher{requester: r}
		got, err := mlbPlayerSearcher.Search(test.pt, 2019, test.playerNamePrefix, test.activePlayersOnly)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: wanted error but did not get one", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		case !reflect.DeepEqual(test.want, got):
			t.Errorf("Test %v: Not equal:\nWanted: %v\nGot:    %v", i, test.want, got)
		}
	}
}

func TestGetMlbPlayerSearchResults(t *testing.T) {
	getPlayerSearchResultsTests := []struct {
		searchResultJSON string
		playerType       db.PlayerType
		wantError        bool
		want             []PlayerSearchResult
	}{
		{
			// bad json
			searchResultJSON: `{}`,
			playerType:       db.PlayerTypeMlbHitter,
		},
		{
			// no results
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"0"}}}`,
			playerType:       db.PlayerTypeMlbHitter,
		},
		{
			// one result
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeMlbHitter,
			want: []PlayerSearchResult{
				{Name: "Mike Trout", Details: "team:LAA, position:CF, born:USA,1991-08-07", SourceID: 545361},
			},
		},
		{
			// two results (multiple results)
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"624413"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
			playerType:       db.PlayerTypeMlbHitter,
			want: []PlayerSearchResult{
				{Name: "Pete Alonso", Details: "team:NYM, position:1B, born:USA,1994-12-07", SourceID: 624413},
				{Name: "Yonder Alonso", Details: "team:COL, position:1B, born:Cuba,1987-04-08", SourceID: 475174},
			},
		},
		{
			// first player_d is invalid
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"INVALID"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
			playerType:       db.PlayerTypeMlbHitter,
			wantError:        true,
		},
		{
			// bad birth_date
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeMlbHitter,
			wantError:        true,
		},
		{
			// no birth_date
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"P","birth_country":"USA","birth_date":"","team_abbrev":"CHC","name_display_first_last":"Abe Johnson","player_id":"116556"}}}}`,
			playerType:       db.PlayerTypeMlbPitcher,
			want: []PlayerSearchResult{
				{Name: "Abe Johnson", Details: "team:CHC, position:P, born:USA,?", SourceID: 116556},
			},
		},
		{
			// no results (wrong playerType)
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeMlbPitcher,
		},
		{
			// no results (wrong playerType)
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeMlbTeam,
		},
	}
	for i, test := range getPlayerSearchResultsTests {
		var mlbPlayerSearchQueryResult MlbPlayerSearch
		err := json.Unmarshal([]byte(test.searchResultJSON), &mlbPlayerSearchQueryResult)
		if err != nil {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err) // (all json should parse into minimum state)
		}
		var got []PlayerSearchResult
		got, err = mlbPlayerSearchQueryResult.SearchPlayerAll.QueryResults.getPlayerSearchResults(test.playerType)
		switch {
		case test.wantError:
			if err == nil {
				t.Errorf("Test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("Test %v: %v", i, err)
		default:
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Test %v:\nwanted: %v\ngot:    %v", i, test.want, got)
			}
		}
	}
}
