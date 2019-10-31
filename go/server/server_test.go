package server

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type mockServerDatastore struct {
	GetYearsFunc func(st db.SportType) ([]db.Year, error)
	adminDatastore
	etlDatastore
}

func (ds mockServerDatastore) GetYears(st db.SportType) ([]db.Year, error) {
	return ds.GetYearsFunc(st)
}

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

func TestNewConfig(t *testing.T) {
	newConfigTests := []struct {
		serverName string
		port       string
		wantErr    bool
	}{
		{ // invalid port
			wantErr: true,
		},
		{ // invalid port
			port:    "four",
			wantErr: true,
		},
		{ // happy path
			serverName: "my server",
			port:       "8000",
		},
	}
	for i, test := range newConfigTests {
		ds := mockServerDatastore{
			nil,
			nil,
			mockEtlDatastore{
				SportTypesFunc: func() db.SportTypeMap {
					return db.SportTypeMap{
						1: {},
						2: {},
					}
				},
			},
		}
		log := log.New(ioutil.Discard, "test", log.LstdFlags)
		cfg, err := NewConfig(test.serverName, ds, test.port, log)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		default:
			if test.serverName != cfg.serverName {
				t.Errorf("Test %v: serverName: wanted %v, got %v", i, test.serverName, cfg.serverName)
			}
			if test.port != cfg.port {
				t.Errorf("Test %v: port: wanted %v, got %v", i, test.port, cfg.port)
			}
			if len(cfg.sportEntries) != 2 {
				t.Errorf("Test %v: wanted len(cfg.sportEntries) to be 2, got %v", i, cfg.sportEntries)
			}
		}
	}
}
