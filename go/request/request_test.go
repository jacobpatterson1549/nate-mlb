package request

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	mockRequestor struct {
		structPointerFromURLFunc func(url string, v interface{}) error
	}

	mockHTTPClient struct {
		DoFunc func(r *http.Request) (*http.Response, error)
	}
)

func (r *mockRequestor) structPointerFromURL(url string, v interface{}) error {
	return r.structPointerFromURLFunc(url, v)
}

func (c mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return c.DoFunc(r)
}

func newMockHTTPRequestor(jsonFunc func(urlPath string) string) requestor {
	return &httpRequestor{
		cache: newCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.WriteString(jsonFunc(r.URL.Path))
				if err != nil {
					return nil, err
				}
				return w.Result(), nil
			},
		},
		logRequestUrls: true,
	}
}

func TestStructPointerFromUrl(t *testing.T) {
	structPointerFromURLTests := []struct {
		url        string
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
			url:       "\x00 (bad url character)",
			wantError: true,
		},
	}
	for i, test := range structPointerFromURLTests {
		jsonFunc := func(urlPath string) string {
			return test.returnJSON
		}
		r := newMockHTTPRequestor(jsonFunc)
		var got interface{}
		err := r.structPointerFromURL(test.url, &got)
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

func TestStructPointerFromUrl_requestorError(t *testing.T) {
	doErr := errors.New("Do error")
	r := httpRequestor{
		cache: newCache(0), // (do not cache)
		httpClient: mockHTTPClient{
			DoFunc: func(r *http.Request) (*http.Response, error) {
				return nil, doErr
			},
		},
		logRequestUrls: true,
	}
	var got interface{}
	err := r.structPointerFromURL("url", &got)
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
func TestStructPointerFromUrl_readBytesError(t *testing.T) {
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
		logRequestUrls: true,
	}
	var got interface{}
	err := r.structPointerFromURL("url", &got)
	if err == nil || !errors.Is(err, readErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}

func TestClearCache(t *testing.T) {
	httpCache = newCache(1)
	url := "url"
	httpCache.add(url, []byte("bytes"))
	if !httpCache.contains(url) {
		t.Error("wanted cache to contain url did not")
	}
	ClearCache()
	if httpCache.contains(url) {
		t.Error("wanted cache to not contain url did")
	}
}
