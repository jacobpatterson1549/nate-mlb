package server

import (
	"testing"
	"time"
)

type previousMidnightTest struct {
	curr time.Time
	want time.Time
}

var pacificLocation, _ = time.LoadLocation("America/Los_Angeles")
var hawaiiLocation, _ = time.LoadLocation("Pacific/Honolulu")
var previousMidnightTests = []previousMidnightTest{
	{
		curr: time.Date(2019, time.August, 22, 0, 0, 0, 0, time.UTC), // 12 AM
		want: time.Date(2019, time.August, 21, 10, 0, 0, 0, time.UTC),
	},
	{
		curr: time.Date(2019, time.August, 21, 16, 15, 17, 0, time.UTC),
		want: time.Date(2019, time.August, 21, 10, 0, 0, 0, time.UTC),
	},
	{
		curr: time.Date(2019, time.August, 22, 2, 0, 0, 0, pacificLocation), // 2 AM
		want: time.Date(2019, time.August, 21, 3, 0, 0, 0, pacificLocation),
	},
	{
		curr: time.Date(2019, time.August, 22, 0, 0, 0, 0, hawaiiLocation), // 12 AM
		want: time.Date(2019, time.August, 22, 0, 0, 0, 0, hawaiiLocation),
	},
}

func TestPreviousMidnight(t *testing.T) {
	for i, test := range previousMidnightTests {
		got := previousMidnight(test.curr)
		if test.want != got {
			t.Errorf("Test %d:\n\twanted %v\n\tgot    %v", i, test.want, got)
		}
	}
}
