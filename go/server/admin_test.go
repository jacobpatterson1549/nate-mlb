package server

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestHandleAdminPostRequest(t *testing.T) {
	handleAdminPostRequestTests := []struct {
		action                   string
		username                 string
		password                 string
		isCorrectUserPassword    bool
		isCorrectUserPasswordErr error
		st                       db.SportType
		wantErr                  bool
		wantCacheCleared         bool
		wantActionCount          int
	}{
		{
			isCorrectUserPassword: false,
			wantErr:               true,
		},
		{
			isCorrectUserPasswordErr: fmt.Errorf("problem checking password"),
			wantErr:                  true,
		},
		{
			isCorrectUserPassword: true,
			action:                "unknown",
			wantErr:               true,
		},
		{
			isCorrectUserPassword: true,
			action:                "friends",
			wantActionCount:       2,
		},
		{
			isCorrectUserPassword: true,
			action:                "players",
			wantActionCount:       2,
		},
		{
			isCorrectUserPassword: true,
			action:                "years",
			wantActionCount:       1, // load previously cached year data if possible
		},
		{
			isCorrectUserPassword: true,
			action:                "cache",
			wantActionCount:       2, // clear cache and stat
		},
		{
			isCorrectUserPassword: true,
			action:                "password",
			wantActionCount:       1,
		},
	}
	for i, test := range handleAdminPostRequestTests {
		ds := mockAdminDatastore{
			IsCorrectUserPasswordFunc: func(username string, p db.Password) (bool, error) {
				return test.isCorrectUserPassword, test.isCorrectUserPasswordErr
			},
		}
		c := mockCache{}
		gotActionCount := 0
		switch test.action {
		case "friends":
			ds.SaveFriendsFunc = func(st db.SportType, futureFriends []db.Friend) error {
				gotActionCount++
				return nil
			}
			ds.ClearStatFunc = func(st db.SportType) error {
				gotActionCount++
				return nil
			}
		case "players":
			ds.SavePlayersFunc = func(st db.SportType, futurePlayers []db.Player) error {
				gotActionCount++
				return nil
			}
			ds.ClearStatFunc = func(st db.SportType) error {
				gotActionCount++
				return nil
			}
		case "years":
			ds.SaveYearsFunc = func(st db.SportType, futureYears []db.Year) error {
				gotActionCount++
				return nil
			}
		case "cache":
			c.ClearFunc = func() {
				gotActionCount++
			}
			ds.ClearStatFunc = func(st db.SportType) error {
				gotActionCount++
				return nil
			}
		case "password":
			ds.SetUserPasswordFunc = func(username string, p db.Password) error {
				gotActionCount++
				return nil
			}
		}
		r := httptest.NewRequest("GET", "http://localhost/admin", nil)
		q := r.URL.Query()
		q.Add("action", test.action)
		r.URL.RawQuery = q.Encode()

		gotErr := handleAdminPostRequest(ds, c, test.st, r)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		default:
			if test.wantActionCount != gotActionCount {
				t.Errorf("Test %v: wanted %v action to run, got %v", i, test.wantActionCount, gotActionCount)
			}
		}
	}
}

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
		r := httptest.NewRequest("GET", "http://localhost/admin/search", nil)
		q := r.URL.Query()
		q.Add("q", test.searchQuery)
		q.Add("pt", test.playerTypeID)
		q.Add("apo", test.activePlayersOnly)
		r.URL.RawQuery = q.Encode()
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

type mockAdminDatastore struct {
	SaveYearsFunc             func(st db.SportType, futureYears []db.Year) error
	SaveFriendsFunc           func(st db.SportType, futureFriends []db.Friend) error
	SavePlayersFunc           func(st db.SportType, futurePlayers []db.Player) error
	ClearStatFunc             func(st db.SportType) error
	SetUserPasswordFunc       func(username string, p db.Password) error
	IsCorrectUserPasswordFunc func(username string, p db.Password) (bool, error)
}

func (ds mockAdminDatastore) SaveYears(st db.SportType, futureYears []db.Year) error {
	return ds.SaveYearsFunc(st, futureYears)
}
func (ds mockAdminDatastore) SaveFriends(st db.SportType, futureFriends []db.Friend) error {
	return ds.SaveFriendsFunc(st, futureFriends)
}
func (ds mockAdminDatastore) SavePlayers(st db.SportType, futurePlayers []db.Player) error {
	return ds.SavePlayersFunc(st, futurePlayers)
}
func (ds mockAdminDatastore) ClearStat(st db.SportType) error {
	return ds.ClearStatFunc(st)
}
func (ds mockAdminDatastore) SetUserPassword(username string, p db.Password) error {
	return ds.SetUserPasswordFunc(username, p)
}
func (ds mockAdminDatastore) IsCorrectUserPassword(username string, p db.Password) (bool, error) {
	return ds.IsCorrectUserPasswordFunc(username, p)
}

type mockCache struct {
	ClearFunc func()
}

func (c mockCache) Clear() {
	c.ClearFunc()
}
