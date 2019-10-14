package server

import (
	"reflect"
	"testing"
	"time"

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

func TestPlayerTypes(t *testing.T) {
	pt1 := db.PlayerType(1)
	pt2 := db.PlayerType(2)
	pt3 := db.PlayerType(3)
	st1 := db.SportType(1)
	st2 := db.SportType(2)
	playerTypes = map[db.PlayerType]db.PlayerTypeInfo{
		pt1: {SportType: st1, DisplayOrder: 2},
		pt2: {SportType: st2, DisplayOrder: 0},
		pt3: {SportType: st1, DisplayOrder: 1},
	}
	want := []db.PlayerType{
		pt3,
		pt1,
	}
	got := getPlayerTypes(st1)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}
