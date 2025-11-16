package request

import (
	"strings"
	"testing"
)

func TestStructPointerFromURI(t *testing.T) {
	tests := map[string]string{
		"":                   "/v2?&appKey=XYZ",
		"path":               "/path?&appKey=XYZ",
		"/path":              "/path?&appKey=XYZ",
		"/path?":             "/path?&appKey=XYZ",
		"path?a=b":           "/path?a=b&appKey=XYZ",
		"/path?a=b":          "/path?a=b&appKey=XYZ",
		"/path/to/x?a=b&c=d": "/path/to/x?a=b&c=d&appKey=XYZ",
	}
	for providedURI, wantURI := range tests {
		jsonFunc := func(gotURI string) string {
			if !strings.HasSuffix(gotURI, wantURI) {
				t.Errorf("when input uri is %v, wanted requested uri to end with %v, but got %v", providedURI, wantURI, gotURI)
			}
			return ""
		}
		r := newMockHTTPRequester(jsonFunc)
		nflR := nflRequester{
			appKey:    "XYZ",
			requester: r,
		}
		nflR.structPointerFromURI(providedURI, nil)
	}
}
