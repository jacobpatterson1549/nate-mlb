package server

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestHandleAdminSearchRequest(t *testing.T) {
	handleAdminSearchRequestTests := []struct {
		searchQuery                 string
		playerTypeID                string
		activePlayersOnly           string
		year                        int
		searcherPlayerSearchResults []request.PlayerSearchResult
		wantSearchQuery             string
		wantPlayerType              db.PlayerType
		wantActivePlayersOnly       bool
		wantPlayerSearchResults     []request.PlayerSearchResult
		wantErr                     bool
	}{
		{ // no searchQuery
			wantErr: true,
		},
		{ // no playerTypeID
			searchQuery: "jose",
			wantErr:     true,
		},
		{ // bad playerType
			searchQuery:  "jose",
			playerTypeID: "one",
			wantErr:      true,
		},
		{ // no searcher for playerType
			searchQuery:  "jose",
			playerTypeID: "1",
			wantErr:      true,
		},
		{ // no searcher for playerType
			searchQuery:  "jose",
			playerTypeID: "2",
			wantErr:      true,
		},
		{ // happy path
			searchQuery:       "jose",
			playerTypeID:      "3",
			year:              2019,
			activePlayersOnly: "off",
			searcherPlayerSearchResults: []request.PlayerSearchResult{
				{Name: "happy path #1"},
			},
			wantPlayerType:        db.PlayerType(3),
			wantActivePlayersOnly: true,
			wantPlayerSearchResults: []request.PlayerSearchResult{
				{Name: "happy path #1"},
			},
		},
		{ // happy path
			searchQuery:       "jose",
			playerTypeID:      "2",
			year:              2001,
			activePlayersOnly: "off",
			searcherPlayerSearchResults: []request.PlayerSearchResult{
				{Name: "happy path #2"},
			},
			wantPlayerType:        db.PlayerType(2),
			wantActivePlayersOnly: false,
			wantPlayerSearchResults: []request.PlayerSearchResult{
				{Name: "happy path #2"},
			},
		},
	}
	for i, test := range handleAdminSearchRequestTests {
		// q := http.UR // TODO: get this working
		r := httptest.NewRequest("GET", "http://localhost/admin/search", nil)
		q := r.URL.Query()
		q.Add("q", test.searchQuery)
		q.Add("pt", test.playerTypeID)
		q.Add("apo", test.activePlayersOnly)
		r.URL.RawQuery = q.Encode()
		// fmt.Println(i, r.URL.String())
		searchers := make(map[db.PlayerType]request.Searcher, 1)
		if len(test.searcherPlayerSearchResults) > 0 {
			searchers[test.wantPlayerType] = mockSearcher{
				SearchFunc: func(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]request.PlayerSearchResult, error) {
					return test.searcherPlayerSearchResults, nil
				},
			}
		}
		gotPlayerSearchResults, gotErr := handleAdminSearchRequest(test.year, searchers, r)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case !reflect.DeepEqual(test.wantPlayerSearchResults, gotPlayerSearchResults):
			t.Errorf("Test %v: not equal:\nwant: %v\ngot:  %v", i, test.wantPlayerSearchResults, gotPlayerSearchResults)
		}
	}
}

type mockSearcher struct {
	SearchFunc func(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]request.PlayerSearchResult, error)
}

func (s mockSearcher) Search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]request.PlayerSearchResult, error) {
	return s.SearchFunc(pt, year, playerNamePrefix, activePlayersOnly)
}
