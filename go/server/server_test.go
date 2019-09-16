package server

import (
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type transformURLPathTest struct {
	urlPath       string
	wantSportType db.SportType
	wantURLPath   string
}

var transformURLPathTests = []transformURLPathTest{
	{
		urlPath:       "",
		wantSportType: 0,
		wantURLPath:   "",
	},
	{
		urlPath:       "/",
		wantSportType: 0,
		wantURLPath:   "",
	},
	{
		urlPath:       "/mlb",
		wantSportType: 1,
		wantURLPath:   "/SportType",
	},
	{
		urlPath:       "/nfl/admin",
		wantSportType: 2,
		wantURLPath:   "/SportType/nfl/admin",
	},
	{
		urlPath:       "/admin",
		wantSportType: 0,
		wantURLPath:   "admin",
	},
}

func TestTransformURLPath(t *testing.T) {
	mockSportTypes := map[string]db.SportType{
		"mlb": 1,
		"nfl": 2,
	}
	var mockSTUR sportTypeURLResolver = func(urlPath string) db.SportType {
		return mockSportTypes[urlPath]
	}
	for i, test := range transformURLPathTests {
		gotSportType, gotURLPath := transformURLPath(test.urlPath, mockSTUR)
		if test.wantSportType != gotSportType || test.wantURLPath != test.wantURLPath {
			t.Errorf("Test %d: wanted '{%v,%v}', but got '{%v,%v}' for url '%v'", i, test.wantSportType, test.wantURLPath, gotSportType, gotURLPath, test.urlPath)
		}
	}
}
