package server

import (
	"errors"
	"net/http/httptest"
	"reflect"
	"sort"
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
			isCorrectUserPasswordErr: errors.New("problem checking password"),
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
		r := httptest.NewRequest("POST", "/admin", nil)
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

func TestUpdateFriends(t *testing.T) {
	updateFriendsTests := []struct {
		st              db.SportType
		form            map[string][]string
		saveErr         error
		wantErr         bool
		wantSaveFriends []db.Friend
	}{
		{},
		{
			saveErr: errors.New("save friends error"),
		},
		{ // bad displayOrder
			form: map[string][]string{
				"friend-8-display-order": {"ONE"},
				"friend-8-name":          {"bart"},
			},
			wantErr: true,
		},
		{ // bad friendId (large string is ok)
			form: map[string][]string{
				"friend-1234567890123456789012345678901234567890-display-order": {"1"},
				"friend-1234567890123456789012345678901234567890-name":          {"bart"},
			},
			wantSaveFriends: []db.Friend{
				{
					ID:           "1234567890123456789012345678901234567890",
					DisplayOrder: 1,
					Name:         "bart",
				},
			},
		},
		{ // happy path
			form: map[string][]string{
				"friend-8-display-order":   {"2"},
				"friend-007-display-order": {"1"},
				"friend-8-name":            {"bart"},
				"friend-007-name":          {"alf"},
			},
			wantSaveFriends: []db.Friend{
				{
					ID:           "007",
					DisplayOrder: 1,
					Name:         "alf",
				},
				{
					ID:           "8",
					DisplayOrder: 2,
					Name:         "bart",
				},
			},
		},
	}
	for i, test := range updateFriendsTests {
		ds := mockAdminDatastore{
			SaveFriendsFunc: func(st db.SportType, futureFriends []db.Friend) error {
				friendDisplayOrder := func(i int) int {
					return futureFriends[i].DisplayOrder
				}
				sort.Slice(futureFriends, func(i, j int) bool {
					return friendDisplayOrder(i) < friendDisplayOrder(j)
				})
				if !reflect.DeepEqual(test.wantSaveFriends, futureFriends) {
					t.Errorf("Test %v:\nwanted save friends: %v\ngot: %v", i, test.wantSaveFriends, futureFriends)
				}
				return test.saveErr
			},
		}
		if test.saveErr == nil {
			ds.ClearStatFunc = func(st db.SportType) error {
				return nil
			}
		}
		r := httptest.NewRequest("POST", "/admin", nil)
		q := r.URL.Query()
		for key, values := range test.form {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		r.URL.RawQuery = q.Encode()
		if err := r.ParseForm(); err != nil {
			t.Errorf("Test %v: could not parse request form: %v", i, err)
		}
		gotErr := updateFriends(ds, test.st, r)
		switch {
		case test.saveErr != nil:
			if !errors.Is(gotErr, test.saveErr) {
				t.Errorf("Test %v: wanted error %v, bug got %v", i, test.saveErr, gotErr)
			}
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestUpdatePlayers(t *testing.T) {
	updatePlayersTests := []struct {
		st              db.SportType
		form            map[string][]string
		saveErr         error
		wantErr         bool
		wantSavePlayers []db.Player
	}{
		{},
		{
			saveErr: errors.New("save players error"),
		},
		{ // bad displayOrder
			form: map[string][]string{
				"player-7-display-order": {"ONE"},
				"player-7-player-type":   {"3"},
				"player-7-friend-id":     {"9"},
				"player-7-source-id":     {"8"},
			},
			wantErr: true,
		},
		{ // large playerId is ok
			form: map[string][]string{
				"player-1234567890123456789012345678901234567890-display-order": {"1"},
				"player-1234567890123456789012345678901234567890-player-type":   {"3"},
				"player-1234567890123456789012345678901234567890-friend-id":     {"9"},
				"player-1234567890123456789012345678901234567890-source-id":     {"8"},
			},
			wantSavePlayers: []db.Player{
				{
					ID:           "1234567890123456789012345678901234567890",
					DisplayOrder: 1,
					PlayerType:   3,
					SourceID:     8,
					FriendID:     "9",
				},
			},
		},
		{ // string friendId is ok
			form: map[string][]string{
				"player-7-display-order": {"1"},
				"player-7-player-type":   {"3"},
				"player-7-friend-id":     {"9x7"},
				"player-7-source-id":     {"8"},
			},
			wantSavePlayers: []db.Player{
				{
					ID:           "7",
					DisplayOrder: 1,
					PlayerType:   3,
					SourceID:     8,
					FriendID:     "9x7",
				},
			},
		},
		{ // bad sourceId
			form: map[string][]string{
				"player-7-display-order": {"1"},
				"player-7-player-type":   {"3"},
				"player-7-friend-id":     {"9"},
				"player-7-source-id":     {"."},
			},
			wantErr: true,
		},
		{ // bad playerType
			form: map[string][]string{
				"player-7-display-order": {"1"},
				"player-7-player-type":   {"low"},
				"player-7-friend-id":     {"9"},
				"player-7-source-id":     {"8"},
			},
			wantErr: true,
		},
		{ // happy path
			form: map[string][]string{
				"player-6-display-order": {"2"},
				"player-7-display-order": {"1"},
				"player-6-player-type":   {"1"},
				"player-7-player-type":   {"2"},
				"player-6-friend-id":     {"4"},
				"player-7-friend-id":     {"6"},
				"player-6-source-id":     {"8000"},
				"player-7-source-id":     {"47"},
			},
			wantSavePlayers: []db.Player{
				{
					ID:           "7",
					DisplayOrder: 1,
					PlayerType:   2,
					SourceID:     47,
					FriendID:     "6",
				},
				{
					ID:           "6",
					DisplayOrder: 2,
					PlayerType:   1,
					SourceID:     8000,
					FriendID:     "4",
				},
			},
		},
	}
	for i, test := range updatePlayersTests {
		ds := mockAdminDatastore{
			SavePlayersFunc: func(st db.SportType, futurePlayers []db.Player) error {
				playerDisplayOrder := func(i int) int {
					return futurePlayers[i].DisplayOrder
				}
				sort.Slice(futurePlayers, func(i, j int) bool {
					return playerDisplayOrder(i) < playerDisplayOrder(j)
				})
				if !reflect.DeepEqual(test.wantSavePlayers, futurePlayers) {
					t.Errorf("Test %v:\nwanted save players: %v\ngot: %v", i, test.wantSavePlayers, futurePlayers)
				}
				return test.saveErr
			},
		}
		if test.saveErr == nil {
			ds.ClearStatFunc = func(st db.SportType) error {
				return nil
			}
		}
		r := httptest.NewRequest("POST", "/admin", nil)
		q := r.URL.Query()
		for key, values := range test.form {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		r.URL.RawQuery = q.Encode()
		if err := r.ParseForm(); err != nil {
			t.Errorf("Test %v: could not parse request form: %v", i, err)
		}
		gotErr := updatePlayers(ds, test.st, r)
		switch {
		case test.saveErr != nil:
			if !errors.Is(gotErr, test.saveErr) {
				t.Errorf("Test %v: wanted error %v, bug got %v", i, test.saveErr, gotErr)
			}
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestUpdateYears(t *testing.T) {
	updateYearsTests := []struct {
		st            db.SportType
		form          map[string][]string
		wantErr       bool
		wantSaveYears []db.Year
	}{
		{},
		{ // bad year
			form: map[string][]string{
				"year": {
					"two thousand nineteen",
				},
			},
			wantErr: true,
		},
		{ // happy path
			form: map[string][]string{
				"year": {
					"2020",
					"2019",
					"2001",
				},
				"year-active": {"2019"},
			},
			wantSaveYears: []db.Year{
				{
					Value:  2020,
					Active: false,
				},
				{
					Value:  2019,
					Active: true,
				},
				{
					Value:  2001,
					Active: false,
				},
			},
		},
	}
	for i, test := range updateYearsTests {
		ds := mockAdminDatastore{
			SaveYearsFunc: func(st db.SportType, futureYears []db.Year) error {
				if !reflect.DeepEqual(test.wantSaveYears, futureYears) {
					t.Errorf("Test %v:\nwanted save years: %v\ngot: %v", i, test.wantSaveYears, futureYears)
				}
				return nil
			},
		}
		r := httptest.NewRequest("POST", "/admin", nil)
		q := r.URL.Query()
		for key, values := range test.form {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		r.URL.RawQuery = q.Encode()
		if err := r.ParseForm(); err != nil {
			t.Errorf("Test %v: could not parse request form: %v", i, err)
		}
		gotErr := updateYears(ds, test.st, r)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestResetPassword(t *testing.T) {
	wantUsername := "fred"
	wantPassword := "s3cr3t&#"
	wantErr := errors.New("password reset error")
	r := httptest.NewRequest("POST", "/admin", nil)
	q := r.URL.Query()
	q.Add("username", wantUsername)
	q.Add("newPassword", wantPassword)
	r.URL.RawQuery = q.Encode()
	ds := mockAdminDatastore{
		SetUserPasswordFunc: func(username string, p db.Password) error {
			if wantUsername != username {
				t.Errorf("wanted %v, got %v", wantUsername, username)
			}
			if wantPassword != string(p) {
				t.Errorf("wanted %v, got %v", wantPassword, p)
			}
			return wantErr
		},
	}
	gotErr := resetPassword(ds, 0, r)
	if wantErr != gotErr {
		t.Errorf("wanted %v, got %v", wantErr, gotErr)
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
