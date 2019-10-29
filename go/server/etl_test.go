package server

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestPreviousMidnight(t *testing.T) {
	pacificLocation, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	hawaiiLocation, err := time.LoadLocation("Pacific/Honolulu")
	if err != nil {
		t.Fatal(err)
	}
	previousMidnightTests := []struct {
		dateTime time.Time
		want     time.Time
	}{
		{
			dateTime: time.Date(2019, time.August, 22, 0, 0, 0, 0, time.UTC), // 12 AM
			want:     time.Date(2019, time.August, 21, 10, 0, 0, 0, time.UTC),
		},
		{
			dateTime: time.Date(2019, time.August, 21, 16, 15, 17, 0, time.UTC),
			want:     time.Date(2019, time.August, 21, 10, 0, 0, 0, time.UTC),
		},
		{
			dateTime: time.Date(2019, time.August, 22, 2, 0, 0, 0, pacificLocation), // 2 AM
			want:     time.Date(2019, time.August, 21, 3, 0, 0, 0, pacificLocation),
		},
		{
			dateTime: time.Date(2019, time.August, 22, 0, 0, 0, 0, hawaiiLocation), // 12 AM
			want:     time.Date(2019, time.August, 22, 0, 0, 0, 0, hawaiiLocation),
		},
	}
	for i, test := range previousMidnightTests {
		got := previousMidnight(test.dateTime)
		if test.want != got {
			t.Errorf("Test %d:\n\twanted %v\n\tgot    %v", i, test.want, got)
		}
	}
}

type mockEtlDatastore struct {
	GetStatFunc     func(st db.SportType) (*db.Stat, error)
	GetFriendsFunc  func(st db.SportType) ([]db.Friend, error)
	GetPlayersFunc  func(st db.SportType) ([]db.Player, error)
	SetStatFunc     func(stat db.Stat) error
	SportTypesFunc  func() db.SportTypeMap
	PlayerTypesFunc func() db.PlayerTypeMap
	GetUtcTimeFunc  func() time.Time
}

func (m mockEtlDatastore) GetStat(st db.SportType) (*db.Stat, error) {
	return m.GetStatFunc(st)
}
func (m mockEtlDatastore) GetFriends(st db.SportType) ([]db.Friend, error) {
	return m.GetFriendsFunc(st)
}
func (m mockEtlDatastore) GetPlayers(st db.SportType) ([]db.Player, error) {
	return m.GetPlayersFunc(st)
}
func (m mockEtlDatastore) SetStat(stat db.Stat) error {
	return m.SetStatFunc(stat)
}
func (m mockEtlDatastore) SportTypes() db.SportTypeMap {
	return m.SportTypesFunc()
}
func (m mockEtlDatastore) PlayerTypes() db.PlayerTypeMap {
	return m.PlayerTypesFunc()
}
func (m mockEtlDatastore) GetUtcTime() time.Time {
	return m.GetUtcTimeFunc()
}

func TestGetScoreCategory(t *testing.T) {
	getScoreCategoryTests := []struct {
		pt                db.PlayerType
		pti               db.PlayerTypeInfo
		year              int
		friends           []db.Friend
		players           []db.Player
		scoreCategorizer  request.ScoreCategorizer
		wantErr           bool
		wantScoreCategory request.ScoreCategory
	}{
		{ // no ScoreCategorizer
			wantErr: true,
		},
		{ // happy path
			scoreCategorizer: mockScoreCategorizer{
				RequestScoreCategoryFunc: func(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (request.ScoreCategory, error) {
					return request.ScoreCategory{Name: "points"}, nil
				},
			},
			wantScoreCategory: request.ScoreCategory{Name: "points"},
		},
		{ // problem requesting score category
			scoreCategorizer: mockScoreCategorizer{
				RequestScoreCategoryFunc: func(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (request.ScoreCategory, error) {
					return request.ScoreCategory{}, fmt.Errorf("request error")
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getScoreCategoryTests {
		scoreCategories := make(chan request.ScoreCategory, 1)
		quit := make(chan error, 1)
		sci := scoreCategoryInfo{
			pt:      test.pt,
			pti:     test.pti,
			year:    test.year,
			friends: test.friends,
			players: test.players,
		}
		getScoreCategory(sci, test.scoreCategorizer, scoreCategories, quit)
		select {
		case got := <-scoreCategories:
			if test.wantErr || !reflect.DeepEqual(test.wantScoreCategory, got) {
				t.Errorf("Test %v: wanted scoreCategory %v, got scoreCategory: %v (expected error: %v)", i, test.wantScoreCategory, got, test.wantErr)
			}
		case got := <-quit:
			if !test.wantErr {
				t.Errorf("Test %v: unexpected error: %v", i, got)
			}
		default:
			t.Errorf("Test %v: did not get message on any channel", i)
		}
	}
}

type mockScoreCategorizer struct {
	RequestScoreCategoryFunc func(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (request.ScoreCategory, error)
}

func (m mockScoreCategorizer) RequestScoreCategory(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (request.ScoreCategory, error) {
	return m.RequestScoreCategoryFunc(pt, ptInfo, year, friends, players)
}

func TestGetPlayerTypes(t *testing.T) {
	pt1 := db.PlayerType(1)
	pt2 := db.PlayerType(2)
	pt3 := db.PlayerType(3)
	st1 := db.SportType(1)
	st2 := db.SportType(2)
	playerTypes := db.PlayerTypeMap{
		pt1: {SportType: st1, DisplayOrder: 2},
		pt2: {SportType: st2, DisplayOrder: 0},
		pt3: {SportType: st1, DisplayOrder: 1},
	}
	want := []db.PlayerType{
		pt3,
		pt1,
	}
	got := getPlayerTypes(st1, playerTypes)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestSetStat(t *testing.T) {
	time1 := time.Date(2019, time.October, 17, 15, 41, 42, 0, time.UTC)
	setStatTests := []struct {
		statEtlTimestamp time.Time
		statEtlJSON      []byte
		wantErr          bool
		want             EtlStats
	}{
		{ // no EtlJSON (see below for set iff len>0 switch)
			wantErr: true,
		},
		{ // bad EtlJSON
			statEtlJSON: []byte(`bad encoding`),
			wantErr:     true,
		},
		{ // happy path
			statEtlTimestamp: time1,
			statEtlJSON:      []byte(`[{"Name":"something"},{"Name":"misc","Description":"?"}]`),
			want: EtlStats{
				etlTime: time1,
				scoreCategories: []request.ScoreCategory{
					{Name: "something"},
					{Name: "misc", Description: "?"},
				},
			},
		},
	}
	for i, test := range setStatTests {
		es := EtlStats{}
		stat := db.Stat{
			EtlTimestamp: &test.statEtlTimestamp,
		}
		if len(test.statEtlJSON) > 0 {
			stat.EtlJSON = &test.statEtlJSON
		}
		gotErr := es.setStat(stat)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case !reflect.DeepEqual(test.want, es):
			t.Errorf("Test %v: did not mutate caller correctly:\n wanted: %v\ngot:    %v", i, test.want, es)
		}
	}
}
