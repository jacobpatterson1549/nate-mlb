package request

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	mockRequestor struct {
		structPointerFromURLFunc func(url string, v interface{}) error
	}

	mockHTTPClient struct {
		JSONFunc func(urlPath string) string
	}
)

func (m *mockRequestor) structPointerFromURL(url string, v interface{}) error {
	return m.structPointerFromURLFunc(url, v)
}

func newMockRequestor(jsonFunc func(urlPath string) string) requestor {
	return &httpRequestor{
		cache:          newCache(0), // (do not cache)
		httpClient:     mockHTTPClient{JSONFunc: jsonFunc},
		logRequestUrls: true,
	}
}

func (m mockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	_, err := w.WriteString(m.JSONFunc(r.URL.Path))
	if err != nil {
		return nil, err
	}
	return w.Result(), nil
}

type structPointerFromURLTest struct {
	url        string
	returnJSON string
	wantError  bool
	want       interface{}
}

var structPointerFromURLTests = []structPointerFromURLTest{
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

func TestStructPointerFromUrl(t *testing.T) {
	for i, test := range structPointerFromURLTests {
		jsonFunc := func(urlPath string) string {
			return test.returnJSON
		}
		r := newMockRequestor(jsonFunc)
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

type httpClientDoError struct {
	doErr error
}

func (c httpClientDoError) Do(r *http.Request) (*http.Response, error) {
	return nil, c.doErr
}

func TestStructPointerFromUrl_requestorError(t *testing.T) {
	doErr := errors.New("Do error")
	r := httpRequestor{
		cache:          newCache(0), // (do not cache)
		httpClient:     httpClientDoError{doErr: doErr},
		logRequestUrls: true,
	}
	var got interface{}
	err := r.structPointerFromURL("url", &got)
	if err == nil || !errors.Is(err, doErr) {
		t.Errorf("expected request to fail, but did not or got wrong error: %v", err)
	}
}
