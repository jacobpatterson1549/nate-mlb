package server

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/request"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
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

func TestGetScoreCategory(t *testing.T) {
	getScoreCategoryTests := []struct {
		pt                   db.PlayerType
		pti                  db.PlayerTypeInfo
		year                 int
		friends              []db.Friend
		players              []db.Player
		requestScoreCategory request.ScoreCategory
		requestErr           error
	}{
		{},
		{
			requestScoreCategory: request.ScoreCategory{Name: "points"},
		},
		{
			requestErr: fmt.Errorf("request error"),
		},
	}
	for i, test := range getScoreCategoryTests {
		scoreCategories := make(chan request.ScoreCategory, 1)
		quit := make(chan error, 1)
		scoreCategorizer := mockScoreCategorizer{
			RequestScoreCategoryFunc: func(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (request.ScoreCategory, error) {
				return test.requestScoreCategory, test.requestErr
			},
		}
		getScoreCategory(test.pt, test.pti, test.year, test.friends, test.players, scoreCategorizer, scoreCategories, quit)
		wantErr := test.requestErr != nil
		select {
		case got := <-scoreCategories:
			if !reflect.DeepEqual(test.requestScoreCategory, got) || wantErr {
				t.Errorf("Test %v: wanted scoreCategory %v, got scoreCategory: %v (expected error: %v", i, test.requestScoreCategory, got, wantErr)
			}
		case got := <-quit:
			if !wantErr {
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
