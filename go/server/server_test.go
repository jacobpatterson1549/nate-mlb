package server

import "testing"

type firstPathSegmentTest struct {
	urlPath string
	want    string
}

var firstPathSegmentTests = []firstPathSegmentTest{
	{
		urlPath: "",
		want:    "",
	},
	{
		urlPath: "/",
		want:    "",
	},
	{
		urlPath: "/mlb",
		want:    "mlb",
	},
	{
		urlPath: "/nfl/admin",
		want:    "nfl",
	},
	{
		urlPath: "/admin",
		want:    "admin", // not a valid sportName, but is still the first path segment
	},
}

func TestFirstPathSegment(t *testing.T) {
	for i, test := range firstPathSegmentTests {
		got := getFirstPathSegment(test.urlPath)
		if test.want != got {
			t.Errorf("Test %d: wanted '%v', but got '%v' for url '%v'", i, test.want, got, test.urlPath)
		}
	}
}
