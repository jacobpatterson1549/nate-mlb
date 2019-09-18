package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	requestor interface {
		structPointerFromURL(url string, v interface{}) error
	}

	httpClient interface {
		Do(r *http.Request) (*http.Response, error)
	}

	httpRequestor struct {
		cache          cache
		httpClient     httpClient
		logRequestUrls bool
	}
)

// ScoreCategorizers maps PlayerTypes to ScoreCategorizers for them
var ScoreCategorizers = make(map[db.PlayerType]ScoreCategorizer)

// Searchers maps PlayerTypes to Searchers for them.
var Searchers = make(map[db.PlayerType]searcher)

// About provides details about the deployment of the application
var About aboutRequestor

var httpCache cache

func init() {
	c := newCache(100)
	r := httpRequestor{
		cache: c,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logRequestUrls: false,
	}

	mlbTeamRequestor := mlbTeamRequestor{requestor: &r}
	mlbPlayerScoreCategorizer := mlbPlayerRequestor{requestor: &r}
	mlbPlayerSearcher := mlbPlayerSearcher{requestor: &r}
	nflTeamRequestor := nflTeamRequestor{requestor: &r}
	nflPlayerRequestor := nflPlayerRequestor{requestor: &r}

	ScoreCategorizers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	ScoreCategorizers[db.PlayerTypeHitter] = &mlbPlayerScoreCategorizer
	ScoreCategorizers[db.PlayerTypePitcher] = &mlbPlayerScoreCategorizer
	ScoreCategorizers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	ScoreCategorizers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	ScoreCategorizers[db.PlayerTypeNflMisc] = &nflPlayerRequestor

	Searchers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	Searchers[db.PlayerTypeHitter] = &mlbPlayerSearcher
	Searchers[db.PlayerTypePitcher] = &mlbPlayerSearcher
	Searchers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	Searchers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	Searchers[db.PlayerTypeNflMisc] = &nflPlayerRequestor
}

func (r *httpRequestor) structPointerFromURL(url string, v interface{}) error {
	var (
		b   []byte
		err error
		ok  bool
	)
	if b, ok = r.cache.get(url); !ok {
		b, err = r.bytes(url)
		if err != nil {
			return err
		}
		r.cache.add(url, b)
	}
	json.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("reading json when requesting %v: %w", url, err)
	}
	return nil
}

func (r *httpRequestor) bytes(url string) ([]byte, error) {
	if r.logRequestUrls {
		log.Printf("%T : requesting %v", r.httpClient, url)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("initializing request to %v: %w", url, err)
	}
	request.Header.Add("Accept", "application/json")
	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting %v: %w", url, err)
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body of %v", url)
	}
	return b, nil
}

// ClearCache clears the request cache
func ClearCache() {
	httpCache.clear()
}
