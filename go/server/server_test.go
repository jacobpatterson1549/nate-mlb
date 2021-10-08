package server

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

type mockHTTPClient struct {
	DoFunc func(r *http.Request) (*http.Response, error)
}

func (m mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return m.DoFunc(r)
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
		htmlFS := fstest.MapFS{}
		jsFS := fstest.MapFS{}
		staticFS := fstest.MapFS{}
		httpClient := mockHTTPClient{}
		cfg, err := NewConfig(test.serverName, ds, test.port, "dummyNflAppKey", logRequestURIs, log, htmlFS, jsFS, staticFS, httpClient)
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
			if !reflect.DeepEqual(htmlFS, cfg.htmlFS) {
				t.Errorf("Test %v: wanted html fs %v, got %v", i, &htmlFS, &cfg.htmlFS)
			}
			if !reflect.DeepEqual(jsFS, cfg.htmlFS) {
				t.Errorf("Test %v: wanted js fs %v, got %v", i, &jsFS, &cfg.jsFS)
			}
			if !reflect.DeepEqual(staticFS, cfg.staticFS) {
				t.Errorf("Test %v: wanted static fs %v, got %v", i, &staticFS, &cfg.staticFS)
			}
		}
	}
}

func TestServerHandlers(t *testing.T) {
	if testing.Short() {
		t.Skip("uses networking")
	}
	tests := []struct {
		wantCode int
		method   string
		path     string
		body     string
	}{
		{wantCode: 200, method: "GET", path: "/"},
		{wantCode: 200, method: "GET", path: "/favicon.ico"},
		{wantCode: 200, method: "GET", path: "/robots.txt"},
		{wantCode: 404, method: "GET", path: "/main.css"},
		{wantCode: 200, method: "GET", path: "/about", body: `[]`},
		{wantCode: 200, method: "GET", path: "/st_1_url"},
		{wantCode: 200, method: "GET", path: "/st_1_url/export"},
		{wantCode: 200, method: "GET", path: "/st_1_url/admin"},
		{wantCode: 200, method: "GET", path: "/st_1_url/admin/search?q=name&pt=1", body: `{}`},
		// {wantCode: 303, method: "POST", path: "/st_1_url/admin?action=password"}, // TODO: add Location header to response
	}
	for i, test := range tests {
		ds := mockServerDatastore{
			GetYearsFunc: func(st db.SportType) ([]db.Year, error) {
				return nil, nil
			},
			adminDatastore: mockAdminDatastore{
				IsCorrectUserPasswordFunc: func(username string, p db.Password) (bool, error) {
					return true, nil
				},
				SetUserPasswordFunc: func(username string, p db.Password) error {
					return nil
				},
			},
			etlDatastore: mockEtlDatastore{
				GetStatFunc: func(st db.SportType) (*db.Stat, error) {
					return nil, nil
				},
				GetFriendsFunc: func(st db.SportType) ([]db.Friend, error) {
					return nil, nil
				},
				GetPlayersFunc: func(st db.SportType) ([]db.Player, error) {
					return nil, nil
				},
				PlayerTypesFunc: func() db.PlayerTypeMap {
					return nil
				},
				SportTypesFunc: func() db.SportTypeMap {
					return db.SportTypeMap{
						1: {URL: "st_1_url"},
					}
				},
				GetUtcTimeFunc: func() time.Time {
					return time.Time{}
				},
			},
		}
		port := "0"
		logRequestURIs := false
		log := log.New(ioutil.Discard, "test", log.LstdFlags)
		htmlFS := fstest.MapFS{
			"html/main/main.html": &fstest.MapFile{
				Data: []byte(
					`<html>
	<head>
		<style>{{template "main.css"}}</style>
	</head>
	<body>
		<p>Tab {{ template "tab.html" }}</p>
	</body>
</html>`),
			},
			"html/home/tab.html":  &fstest.MapFile{Data: []byte(`1`)},
			"html/about/tab.html": &fstest.MapFile{Data: []byte(`2`)},
			"html/stats/tab.html": &fstest.MapFile{Data: []byte(`3`)},
			"html/admin/tab.html": &fstest.MapFile{Data: []byte(`4`)},
		}
		jsFS := fstest.MapFS{}
		staticFS := fstest.MapFS{
			"static/main.css": &fstest.MapFile{
				Data: []byte(`html{color:red;}`),
			},
			"static/favicon.ico": &fstest.MapFile{
				Data: []byte(`[binary-icon]`),
			},
			"static/robots.txt": &fstest.MapFile{
				Data: []byte(``),
			},
		}
		httpClient := mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				resp := http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}
				return &resp, nil
			},
		}
		cfg, _ := NewConfig("serverName", ds, port, "dummyNflAppKey", logRequestURIs, log, htmlFS, jsFS, staticFS, httpClient)
		h := cfg.handler()
		ts := httptest.NewTLSServer(h)
		defer ts.Close()
		client := ts.Client()
		path := ts.URL + test.path
		var resp *http.Response
		var err error
		switch test.method {
		case "GET":
			resp, err = client.Get(path)
		case "POST":
			resp, err = client.Post(path, "dummy_content_type", nil)
		default:
			t.Errorf("Test %v: unwanted method: %v", i, test.method)
		}
		switch {
		case err != nil:
			t.Errorf("Test %v: unwanted error requesting %v: %v", i, path, err)
		case resp.StatusCode != test.wantCode:
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Errorf("Test %v: wanted %v, got %v:\nresponse: %v\nbody read error: %v\nbody: %s", i, test.wantCode, resp.StatusCode, resp, err, body)
		}
	}
}
