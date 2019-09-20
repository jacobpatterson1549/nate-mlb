package request

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestGetPlayerSearchResults(t *testing.T) {
	getPlayerSearchResultsTests := []struct {
		searchResultJSON string
		playerType       db.PlayerType
		wantError        bool
		want             []PlayerSearchResult
	}{
		{
			// bad json
			searchResultJSON: `{}`,
			playerType:       db.PlayerTypeHitter,
		},
		{
			// no results
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"0"}}}`,
			playerType:       db.PlayerTypeHitter,
		},
		{
			// one result
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeHitter,
			want: []PlayerSearchResult{
				{Name: "Mike Trout", Details: "team:LAA, position:CF, born:USA,1991-08-07", SourceID: 545361},
			},
		},
		{
			// two results (multiple results)
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"624413"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
			playerType:       db.PlayerTypeHitter,
			want: []PlayerSearchResult{
				{Name: "Pete Alonso", Details: "team:NYM, position:1B, born:USA,1994-12-07", SourceID: 624413},
				{Name: "Yonder Alonso", Details: "team:COL, position:1B, born:Cuba,1987-04-08", SourceID: 475174},
			},
		},
		{
			// first player_d is invalid
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"INVALID"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
			playerType:       db.PlayerTypeHitter,
			wantError:        true,
		},
		{
			// bad birth_date
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypeHitter,
			wantError:        true,
		},
		{
			// no birth_date
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"P","birth_country":"USA","birth_date":"","team_abbrev":"CHC","name_display_first_last":"Abe Johnson","player_id":"116556"}}}}`,
			playerType:       db.PlayerTypePitcher,
			want: []PlayerSearchResult{
				{Name: "Abe Johnson", Details: "team:CHC, position:P, born:USA,?", SourceID: 116556},
			},
		},
		{
			// no results (wrong playerType)
			searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
			playerType:       db.PlayerTypePitcher,
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
