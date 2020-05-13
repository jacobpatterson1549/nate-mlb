package request

import (
	"fmt"
	"strings"
)

type nflRequester struct {
	appKey    string
	requester requester
}

func (n nflRequester) structPointerFromURI(uri string, v interface{}) error {
	if len(uri) > 0 && uri[0] != '/' {
		uri = "/" + uri
	}
	if !strings.Contains(uri, "?") {
		uri = uri + "?"
	}
	uri = fmt.Sprintf("https://api.fantasy.nfl.com/v2%s&appKey=%s", uri, n.appKey)
	return n.requester.structPointerFromURI(uri, v)
}
