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
	"github.com/jacobpatterson1549/nate-mlb/go/request"
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

type mockAboutRequester struct {
	PreviousDeploymentFunc func() (*request.Deployment, error)
}

func (m mockAboutRequester) PreviousDeployment() (*request.Deployment, error) {
	return m.PreviousDeploymentFunc()
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

	s := Server{
		sportTypesByURL: map[string]db.SportType{
			"mlb": db.SportTypeMlb,
			"nfl": db.SportTypeNfl,
		},
	}
	for i, test := range transformURLPathTests {
		r := httptest.NewRequest("GET", test.urlPath, nil)
		gotSportType, gotURLPath := s.transformURLPath(r)
		switch {
		case test.wantSportType != gotSportType:
			t.Errorf("Test %d: sport types equal for url %v:\nwanted: %v\ngot:    %v", i, test.urlPath, test.wantSportType, gotSportType)
		case test.wantURLPath != gotURLPath:
			t.Errorf("Test %d: urls not equal for url %v:\nwanted: %v\ngot:    %v", i, test.urlPath, test.wantURLPath, gotURLPath)
		}
	}
}

func TestNew(t *testing.T) {
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
		cfg := Config{
			DisplayName:    test.serverName,
			Port:           test.port,
			NflAppKey:      "dummyNflAppKey",
			LogRequestURIs: logRequestURIs,
			HtmlFS:         htmlFS,
			JavascriptFS:   jsFS,
			StaticFS:       staticFS,
		}
		s, err := cfg.New(log, ds, httpClient)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		// happy path testing from here down
		case test.serverName != s.DisplayName:
			t.Errorf("Test %v: serverName: wanted %v, got %v", i, test.serverName, s.DisplayName)
		case cfg.NflAppKey != s.NflAppKey:
			t.Errorf("Test %v: NflAppKey: wanted %v, got %v", i, cfg.NflAppKey, s.NflAppKey)
		case test.port != s.Port:
			t.Errorf("Test %v: port: wanted %v, got %v", i, test.port, s.Port)
		case len(s.sportEntries) != 2:
			t.Errorf("Test %v: wanted len(cfg.sportEntries) to be 2, got %v", i, s.sportEntries)
		case len(s.sportTypesByURL) != 2:
			t.Errorf("Test %v: wanted len(cfg.sportTypesByURL) to be 2, got %v", i, s.sportTypesByURL)
		case s.requestCache == nil:
			t.Errorf("Test %v: cache not set", i)
		case !reflect.DeepEqual(log, s.log):
			t.Errorf("Test %v: wanted log %v, got %v", i, &log, &s.log)
		case s.ds == nil:
			t.Errorf("Test %v: data store not set", i)
		case !reflect.DeepEqual(htmlFS, s.HtmlFS):
			t.Errorf("Test %v: wanted html fs %v, got %v", i, &htmlFS, &s.HtmlFS)
		case !reflect.DeepEqual(jsFS, s.HtmlFS):
			t.Errorf("Test %v: wanted js fs %v, got %v", i, &jsFS, &s.JavascriptFS)
		case !reflect.DeepEqual(staticFS, s.StaticFS):
			t.Errorf("Test %v: wanted static fs %v, got %v", i, &staticFS, &s.StaticFS)
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
	}{
		{wantCode: 200, method: "GET", path: "/"},
		{wantCode: 200, method: "GET", path: "/favicon.ico"},
		{wantCode: 200, method: "GET", path: "/robots.txt"},
		{wantCode: 404, method: "GET", path: "/main.css"},
		{wantCode: 200, method: "GET", path: "/about"},
		{wantCode: 200, method: "GET", path: "/st_1_url"},
		{wantCode: 200, method: "GET", path: "/st_1_url/export"},
		{wantCode: 200, method: "GET", path: "/st_1_url/admin"},
		{wantCode: 200, method: "GET", path: "/st_1_url/admin/search?q=name&pt=77"},
		{wantCode: 200, method: "POST", path: "/st_1_url/admin?action=password"}, // should redirect to 200
		{wantCode: 405, method: "HEAD", path: "/"},
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
		s := Server{
			Config: Config{
				HtmlFS:       htmlFS,
				JavascriptFS: jsFS,
				StaticFS:     staticFS,
			},
			log: log,
			ds:  ds,
			aboutRequester: mockAboutRequester{
				PreviousDeploymentFunc: func() (*request.Deployment, error) {
					return new(request.Deployment), nil
				},
			},
			sportTypesByURL: map[string]db.SportType{
				"st_1_url": 1,
			},
			searchers: map[db.PlayerType]request.Searcher{
				77: mockSearcher{
					SearchFunc: func(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]request.PlayerSearchResult, error) {
						return nil, nil
					},
				},
			},
		}
		h := s.handler()
		ts := httptest.NewTLSServer(h)
		defer ts.Close()
		client := ts.Client()
		path := ts.URL + test.path
		req, _ := http.NewRequest(test.method, path, nil)
		resp, err := client.Do(req)
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

func TestWithGzip(t *testing.T) {
	const wantMessage = "Hello, World!"
	tests := []struct {
		acceptEncoding string
		wantGzip       bool
		wantBodyStart  string
	}{
		{
			wantBodyStart: wantMessage,
		},
		{
			acceptEncoding: "gzip, deflate, br",
			wantGzip:       true,
			wantBodyStart:  "\x1f\x8b\x08", // magic number (1f8b) and compression method for deflate (08)
		},
	}
	for i, test := range tests {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(wantMessage))
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Add("Accept-Encoding", test.acceptEncoding)
		withGzip(h).ServeHTTP(w, r)
		contentEncoding := w.Header().Get("Content-Encoding")
		gotGzip := contentEncoding == "gzip"
		gotMessage := w.Body.String()
		switch {
		case test.wantGzip != gotGzip:
			t.Errorf("Test %v: wanted gzip: %v, got %v", i, test.wantGzip, gotGzip)
		case !strings.HasPrefix(gotMessage, test.wantBodyStart):
			t.Errorf("Test %v: written message prefixes not equal:\nwanted: %x\ngot:    %x", i, test.wantBodyStart, gotMessage)
		}
	}
}
