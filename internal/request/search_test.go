package request

import (
	"encoding/json"
	"fmt"
	"testing"
)

type getPlayerSearchResultsTest struct {
	searchResultJSON string
	wantError        bool
	want             []PlayerSearchResult
}

var getPlayerSearchResultsTests = []getPlayerSearchResultsTest{
	{
		// bad json
		searchResultJSON: `{}`,
		wantError:        true,
	},
	{
		// no results
		searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"0"}}}`},
	{
		// one result
		searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991-08-07T00:00:00","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
		want: []PlayerSearchResult{
			PlayerSearchResult{Name: "Mike Trout", Details: "team:LAA, position:CF, born:USA,1991-08-07", PlayerID: 545361},
		},
	},
	{
		// two results (multiple results)
		searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"624413"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
		want: []PlayerSearchResult{
			PlayerSearchResult{Name: "Pete Alonso", Details: "team:NYM, position:1B, born:USA,1994-12-07", PlayerID: 624413},
			PlayerSearchResult{Name: "Yonder Alonso", Details: "team:COL, position:1B, born:Cuba,1987-04-08", PlayerID: 475174},
		},
	},
	{
		// playerid bad
		searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"2","row":[{"position":"1B","birth_country":"USA","birth_date":"1994-12-07T00:00:00","team_abbrev":"NYM","name_display_first_last":"Pete Alonso","player_id":"INVALID"},{"position":"1B","birth_country":"Cuba","birth_date":"1987-04-08T00:00:00","team_abbrev":"COL","name_display_first_last":"Yonder Alonso","player_id":"475174"}]}}}`,
		wantError:        true,
	},
	{
		// bad birth_date
		searchResultJSON: `{"search_player_all":{"queryResults":{"totalSize":"1","row":{"position":"CF","birth_country":"USA","birth_date":"1991","team_abbrev":"LAA","name_display_first_last":"Mike Trout","player_id":"545361"}}}}`,
		wantError:        true,
	},
}

func TestGetPlayerSearchResults(t *testing.T) {
	for i, test := range getPlayerSearchResultsTests {
		var playerSearchQueryResult PlayerSearchQueryResult
		err := json.Unmarshal([]byte(test.searchResultJSON), &playerSearchQueryResult)
		if err != nil {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err) // (all json should parse into minimum state)
		}
		var got []PlayerSearchResult
		got, err = playerSearchQueryResult.SearchPlayerAll.QueryResults.getPlayerSearchResults()
		hadError := err != nil
		if test.wantError != hadError {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err)
		} else if !test.wantError {
			if err = assertEqualPlayerSearchResults(test.want, got); err != nil {
				t.Errorf("Test %v: %v", i, err)
			}
		}
	}
}

func assertEqualPlayerSearchResults(want, got []PlayerSearchResult) error {
	if len(got) != len(want) {
		return fmt.Errorf("wanted %v, but got different length results %v", want, got)
	}
	for j, w := range want {
		g := got[j]
		if w != g {
			return fmt.Errorf("values at index %v different: wanted %v, but got %v", j, w, g)
		}
	}
	return nil
}
