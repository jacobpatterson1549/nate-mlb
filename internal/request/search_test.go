package request

import (
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
		wantError: true,
	},
	{
		// no results
		searchResultJSON: `{"search_player_all":{"queryResults":{}}}`},
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
}

func TestGetPlayerSearchResults(t *testing.T) {
	for i, test := range getPlayerSearchResultsTests {
		got, err := getPlayerSearchResults([]byte(test.searchResultJSON))
		hadError := err != nil
		if test.wantError != hadError {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err)
		}
		if !test.wantError {
			if len(got) != len(test.want) {
				t.Errorf("Test %v: wanted %v, but got different length results %v", i, test.want, got)
			} else {
				for j, w := range test.want {
					g := got[j]
					if w != g {
						t.Errorf("Test %v: values at index %v different: wanted %v, but got %v", i, j, w, g)
					}
				}
			}
		}
	}
}
