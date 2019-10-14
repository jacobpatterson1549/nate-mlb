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

func TestSportTypeFromURL(t *testing.T) {
	sportTypeFromURLTests := []struct {
		loadedSportTypes map[db.SportType]db.SportTypeInfo
		url              string
		want             db.SportType
	}{
		{},
		{
			loadedSportTypes: map[db.SportType]db.SportTypeInfo{
				db.SportType(2): {URL: "here"},
				db.SportType(8): {URL: "somewhere"},
			},
			url:  "somewhere",
			want: db.SportType(8),
		},
		{
			loadedSportTypes: map[db.SportType]db.SportTypeInfo{
				db.SportType(3): {URL: "*"},
			},
			url:  "anywhere",
			want: db.SportType(0),
		},
	}
	for i, test := range sportTypeFromURLTests {
		sportTypes = test.loadedSportTypes
		got := sportTypeFromURL(test.url)
		if test.want != got {
			t.Errorf("Test :%v:\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}
