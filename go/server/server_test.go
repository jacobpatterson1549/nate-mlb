package server

import (
	"io/ioutil"
	"log"
	"reflect"
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
			urlPath:       "/",
			wantSportType: 0,
			wantURLPath:   "/",
		},
		{
			urlPath:       "/mlb",
			wantSportType: db.SportTypeMlb,
			wantURLPath:   "/SportType",
		},
		{
			urlPath:       "/nfl/admin",
			wantSportType: db.SportTypeNfl,
			wantURLPath:   "/SportType/admin",
		},
		{
			urlPath:       "/admin",
			wantSportType: 0,
			wantURLPath:   "/admin",
		},
	}

	sportTypesURLLookup := map[string]db.SportType{
		"mlb": db.SportTypeMlb,
		"nfl": db.SportTypeNfl,
	}
	for i, test := range transformURLPathTests {
		gotSportType, gotURLPath := transformURLPath(sportTypesURLLookup, test.urlPath)
		switch {
		case test.wantSportType != gotSportType:
			t.Errorf("Test %d: sport types equal for url %v:\nwanted: %v\ngot:    %v", i, test.urlPath, test.wantSportType, gotSportType)
		case test.wantURLPath != gotURLPath:
			t.Errorf("Test %d: urls not equal for url %v:\nwanted: %v\ngot:    %v", i, test.urlPath, test.wantURLPath, gotURLPath)
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
						1: {URL: "st_1_url"},
						2: {URL: "st_2_url"},
					}
				},
			},
		}
		logRequestURIs := false
		log := log.New(ioutil.Discard, "test", log.LstdFlags)
		cfg, err := NewConfig(test.serverName, ds, test.port, "dummyNflAppKey", logRequestURIs, log)
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
			if len(cfg.sportTypesByURL) != 2 {
				t.Errorf("Test %v: wanted len(cfg.sportTypesByURL) to be 2, got %v", i, cfg.sportTypesByURL)
			}
			if cfg.requestCache == nil {
				t.Errorf("Test %v: cache not set", i)
			}
			if !reflect.DeepEqual(log, cfg.log) {
				t.Errorf("Test %v: wanted log %v, got %v", i, &log, &cfg.log)
			}
		}
	}
}
