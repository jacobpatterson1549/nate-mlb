package request

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	mockRequestor struct {
		structPointerFromURIFunc func(uri string, v interface{}) error
	}

	mockHTTPClient struct {
		DoFunc func(r *http.Request) (*http.Response, error)
	}
)

func (r *mockRequestor) structPointerFromURI(uri string, v interface{}) error {
	return r.structPointerFromURIFunc(uri, v)
}

func (c mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return c.DoFunc(r)
}

func newMockHTTPRequestor(jsonFunc func(uriPath string) string) requestor {
	do := func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		uri := r.URL.RequestURI()
		_, err := w.WriteString(jsonFunc(uri))
		if err != nil {
			return nil, err
		}
		return w.Result(), nil
	}
	client := mockHTTPClient{
		DoFunc: do,
	}
	return &httpRequestor{
		cache:      newCache(0), // (do not cache)
		httpClient: client,
		// logRequestUris: true,
	}
}

func TestStructPointerFromUri(t *testing.T) {
	structPointerFromURITests := []struct {
		uri        string
		returnJSON string
		wantError  bool
		want       interface{}
	}{
		{
			returnJSON: `"valid json string"`,
			want:       "valid json string",
		},
		{
			returnJSON: `bad json`,
			wantError:  true,
		},
		{
			uri:       "\x00 (bad uri character)",
			wantError: true,
		},
	}
	for i, test := range structPointerFromURITests {
		jsonFunc := func(uriPath string) string {
			return test.returnJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		var got interface{}
		err := r.structPointerFromURI(test.uri, &got)
		switch {
		case test.wantError:
			if err == nil {
				t.Errorf("Test %d: expected request to fail, but did not", i)
			}
		case test.want != got:
			t.Errorf("Test %d:wanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

func TestStructPointerFromUri_requestorError(t *testing.T) {
	doErr := errors.New("Do error")
	r := httpRequestor{
		cache: newCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				return nil, doErr
			},
		},
		// logRequestUris: true,
	}
	var got interface{}
	err := r.structPointerFromURI("uri", &got)
	if err == nil || !errors.Is(err, doErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}

type mockReadCloser struct {
	io.ReadCloser
	readErr  error
	closeErr error
}

func (m mockReadCloser) Read(b []byte) (n int, err error) {
	return len(b), m.readErr
}
func (m mockReadCloser) Close() error {
	return m.closeErr
}
func TestStructPointerFromUri_readBytesError(t *testing.T) {
	readErr := errors.New("read error")
	r := httpRequestor{
		cache: newCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				response := http.Response{
					Body: mockReadCloser{readErr: readErr},
				}
				return &response, nil
			},
		},
		// logRequestUris: true,
	}
	var got interface{}
	err := r.structPointerFromURI("uri", &got)
	if err == nil || !errors.Is(err, readErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}

func TestClearCache(t *testing.T) {
	httpCache = newCache(1)
	uri := "uri"
	httpCache.add(uri, []byte("bytes"))
	if !httpCache.contains(uri) {
		t.Error("wanted cache to contain uri did not")
	}
	ClearCache()
	if httpCache.contains(uri) {
		t.Error("wanted cache to not contain uri did")
	}
}

type mockScoreCategorizer struct {
	scoreCategory ScoreCategory
}

func (m mockScoreCategorizer) requestScoreCategory(pt db.PlayerType, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	return m.scoreCategory, nil
}

func TestRequestScore(t *testing.T) {
	scoreCategorizers = map[db.PlayerType]scoreCategorizer{
		1: mockScoreCategorizer{scoreCategory: ScoreCategory{Name: "a"}},
		2: mockScoreCategorizer{scoreCategory: ScoreCategory{Name: "b"}},
		3: mockScoreCategorizer{scoreCategory: ScoreCategory{Name: "c"}},
	}

	want := ScoreCategory{Name: "b"}
	got, err := Score(2, 0, nil, nil)

	switch {
	case err != nil:
		t.Error(err)
	case !reflect.DeepEqual(want, got):
		t.Errorf("not equal\nwanted: %v\ngot:    %v", want, got)
	}
}

type mockSearcher struct {
	playerSearchResults []PlayerSearchResult
}

func (m mockSearcher) search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	return m.playerSearchResults, nil
}

func TestPlayerSearchResults(t *testing.T) {
	searchers = map[db.PlayerType]searcher{
		1: mockSearcher{playerSearchResults: []PlayerSearchResult{{Name: "art"}}},
		2: mockSearcher{playerSearchResults: []PlayerSearchResult{{Name: "bart"}}},
		3: mockSearcher{playerSearchResults: []PlayerSearchResult{{Name: "curt"}}},
	}

	want := []PlayerSearchResult{{Name: "curt"}}
	got, err := Search(db.PlayerType(3), 0, "", true)

	switch {
	case err != nil:
		t.Error(err)
	case !reflect.DeepEqual(want, got):
		t.Errorf("not equal\nwanted: %v\ngot:    %v", want, got)
	}
}
