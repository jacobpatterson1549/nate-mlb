package server

import (
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestTransformURLPath(t *testing.T) {
	transformURLPathTests := []struct {
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

	sportTypesURLLookup := map[string]db.SportType{
		"mlb": db.SportTypeMlb,
		"nfl": db.SportTypeNfl,
	}
	for i, test := range transformURLPathTests {
		gotSportType, gotURLPath := transformURLPath(sportTypesURLLookup, test.urlPath)
		if test.wantSportType != gotSportType || test.wantURLPath != test.wantURLPath {
			t.Errorf("Test %d: wanted '{%v,%v}', but got '{%v,%v}' for url '%v'", i, test.wantSportType, test.wantURLPath, gotSportType, gotURLPath, test.urlPath)
		}
	}
}
