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

type mockSportType struct {
	id int
}

func (mst mockSportType) ID() int {
	return mst.id
}

func (mst mockSportType) Name() string {
	return ""
}

func (mst mockSportType) URL() string {
	return ""
}

var mockMlbSportType = mockSportType{id: 1}
var mockNflSportType = mockSportType{id: 2}
var transformURLPathTests = []transformURLPathTest{
	{
		urlPath:       "",
		wantSportType: nil,
		wantURLPath:   "",
	},
	{
		urlPath:       "/",
		wantSportType: nil,
		wantURLPath:   "",
	},
	{
		urlPath:       "/mlb",
		wantSportType: mockMlbSportType,
		wantURLPath:   "/SportType",
	},
	{
		urlPath:       "/nfl/admin",
		wantSportType: mockNflSportType,
		wantURLPath:   "/SportType/nfl/admin",
	},
	{
		urlPath:       "/admin",
		wantSportType: nil,
		wantURLPath:   "admin",
	},
}

func TestTransformURLPath(t *testing.T) {
	mockSportTypes := map[string]db.SportType{
		"mlb": mockMlbSportType,
		"nfl": mockNflSportType,
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
