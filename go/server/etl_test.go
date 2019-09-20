package server

import (
	"testing"
	"time"
)

func TestPreviousMidnight(t *testing.T) {
	pacificLocation, _ := time.LoadLocation("America/Los_Angeles")
	hawaiiLocation, _ := time.LoadLocation("Pacific/Honolulu")
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
