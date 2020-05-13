package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	mockRequester struct {
		structPointerFromURIFunc func(uri string, v interface{}) error
	}

	mockHTTPClient struct {
		DoFunc func(r *http.Request) (*http.Response, error)
	}
)

func (r *mockRequester) structPointerFromURI(uri string, v interface{}) error {
	return r.structPointerFromURIFunc(uri, v)
}

func (c mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return c.DoFunc(r)
}

func newMockHTTPRequester(jsonFunc func(uriPath string) string) requester {
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
	return &httpRequester{
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
		r := newMockHTTPRequester(jsonFunc)
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

func TestHttpRequesterLogRequestUrl(t *testing.T) {
	for _, test := range []bool{true, false} {
		buffer := bytes.NewBufferString("")
		log := log.New(buffer, "test", log.LstdFlags)
		r := httpRequester{
			cache: NewCache(0), // (do not cache)
			httpClient: mockHTTPClient{
				DoFunc: func(r *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("request error")
				},
			},
			logRequestURIs: test,
			log:            log,
		}
		uri := "TEST_URI"
		r.bytes(uri)
		logText := buffer.String()
		if test && !strings.Contains(logText, uri) {
			t.Errorf("expected uri to be written to log, but was not; got: %v", logText)
		}
		if !test && buffer.Len() > 0 {
			t.Errorf("expected nothing to be written to log, but got: %v", logText)
		}
	}
}

func TestStructPointerFromUri_requesterError(t *testing.T) {
	doErr := errors.New("Do error")
	r := httpRequester{
		cache: NewCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				return nil, doErr
			},
		},
	}
	var got interface{}
	err := r.structPointerFromURI("uri", &got)
	if err == nil || !errors.Is(err, doErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}

type mockReadCloser struct {
	readFunc func(b []byte) (n int, err error)
	readErr  error
	closeErr error
}

func (m mockReadCloser) Read(b []byte) (n int, err error) {
	if m.readFunc != nil {
		return m.readFunc(b)
	}
	return len(b), m.readErr
}
func (m mockReadCloser) Close() error {
	return m.closeErr
}

func TestStructPointerFromUri_okRequest(t *testing.T) {
	r := httpRequester{
		cache: NewCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				response := http.Response{
					StatusCode: http.StatusOK,
					Body: mockReadCloser{
						readFunc: func(b []byte) (n int, err error) {
							switch len(b) {
							case 0:
								return -1, nil
							default:
								b[0] = '7'
								return 1, io.EOF
							}
						},
					},
				}
				return &response, nil
			},
		},
	}
	want := 7
	var got int
	err := r.structPointerFromURI("uri", &got)
	switch {
	case err != nil:
		t.Errorf("unexpected error: %v", err)
	case want != got:
		t.Errorf("wanted %d, got %v", want, got)
	}
}

func TestStructPointerFromUri_badRequest(t *testing.T) {
	r := httpRequester{
		cache: NewCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				response := http.Response{
					StatusCode: http.StatusBadRequest, // should cause failure
					Body: mockReadCloser{
						readFunc: func(b []byte) (n int, err error) {
							return -1, nil
						},
					},
				}
				return &response, nil
			},
		},
	}
	err := r.structPointerFromURI("uri", nil)
	if err == nil {
		t.Errorf("expected error because the request is bad")
	}
}

func TestStructPointerFromUri_readBytesError(t *testing.T) {
	readErr := errors.New("read error")
	r := httpRequester{
		cache: NewCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				response := http.Response{
					StatusCode: http.StatusOK,
					Body:       mockReadCloser{readErr: readErr},
				}
				return &response, nil
			},
		},
	}
	var got interface{}
	err := r.structPointerFromURI("uri", &got)
	if err == nil || !errors.Is(err, readErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}

func TestNewRequesters(t *testing.T) {
	c := NewCache(0)
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	scoreCategorizers, searchers, aboutRequester := NewRequesters(c, "dummyNflAppKey", "environmentName", log)
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
	if aboutRequester.environment != "environmentName" {
		t.Errorf("environment not set for aboutRequester")
	}
}
