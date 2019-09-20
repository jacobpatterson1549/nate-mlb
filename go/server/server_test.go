package server

import (
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestTransformURLPath(t *testing.T) {
	var transformURLPathTests = []struct {
		urlPath       string
		wantSportType db.SportType
		wantURLPath   string
	}{
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
			wantSportType: db.SportTypeMlb,
			wantURLPath:   "/SportType",
		},
		{
			urlPath:       "/nfl/admin",
			wantSportType: db.SportTypeNfl,
			wantURLPath:   "/SportType/nfl/admin",
		},
		{
			urlPath:       "/admin",
			wantSportType: 0,
			wantURLPath:   "admin",
		},
	}
	mockSportTypes := map[string]db.SportType{
		"mlb": db.SportTypeMlb,
		"nfl": db.SportTypeNfl,
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
