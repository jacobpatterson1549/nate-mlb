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
		cache:      NewCache(0), // (do not cache)
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
		cache: NewCache(0), // (do not cache)
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
		cache: NewCache(0), // (do not cache)
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

func TestNewRequestors(t *testing.T) {
	c := NewCache(0)
	scoreCategorizers, searchers, aboutRequestor := NewRequestors(c)
	wantPlayerTypes := db.PlayerTypeMap{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}}
	if len(wantPlayerTypes) != len(scoreCategorizers) {
		t.Errorf("expected %v scoreCategorizers, but got %v", len(wantPlayerTypes), len(scoreCategorizers))
	}
	if len(wantPlayerTypes) != len(searchers) {
		t.Errorf("expected %v searchers, but got %v", len(wantPlayerTypes), len(searchers))
	}
	for pt := range wantPlayerTypes {
		if _, ok := scoreCategorizers[pt]; !ok {
			t.Errorf("expected ScoreCategorizer for pt %v", pt)
		}
		if _, ok := searchers[pt]; !ok {
			t.Errorf("expected Searcher for pt %v", pt)
		}
	}
	if !reflect.DeepEqual(aboutRequestor.requestor, c) {
		t.Errorf("requestor not set for aboutRequestor")
	}
}
