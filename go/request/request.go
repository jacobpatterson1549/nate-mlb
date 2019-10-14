// Package request requests scores and performs searches for different sports/players.
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
		structPointerFromURI(uri string, v interface{}) error
	}

	httpClient interface {
		Do(r *http.Request) (*http.Response, error)
	}

	httpRequestor struct {
		cache          cache
		httpClient     httpClient
		logRequestURIs bool
	}
)

var (
	// scoreCategorizers maps PlayerTypes to scoreCategorizers for them
	scoreCategorizers = make(map[db.PlayerType]scoreCategorizer)

	// searchers maps PlayerTypes to searchers for them.
	searchers = make(map[db.PlayerType]searcher)

	// About provides details about the deployment of the application
	About aboutRequestor

	httpCache cache = newCache(100)
)

func init() {
	r := httpRequestor{
		cache: httpCache,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logRequestURIs: false,
	}

	mlbTeamRequestor := mlbTeamRequestor{requestor: &r}
	mlbPlayerScoreCategorizer := mlbPlayerRequestor{requestor: &r}
	mlbPlayerSearcher := mlbPlayerSearcher{requestor: &r}
	nflTeamRequestor := nflTeamRequestor{requestor: &r}
	nflPlayerRequestor := nflPlayerRequestor{requestor: &r}

	scoreCategorizers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	scoreCategorizers[db.PlayerTypeMlbHitter] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeMlbPitcher] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	scoreCategorizers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	scoreCategorizers[db.PlayerTypeNflMisc] = &nflPlayerRequestor

	searchers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	searchers[db.PlayerTypeMlbHitter] = &mlbPlayerSearcher
	searchers[db.PlayerTypeMlbPitcher] = &mlbPlayerSearcher
	searchers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	searchers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	searchers[db.PlayerTypeNflMisc] = &nflPlayerRequestor

	About = aboutRequestor{requestor: &r}
}

// Score gets the ScoreCategory for the PlayerType/year
func Score(pt db.PlayerType, ptInfo db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player) (ScoreCategory, error) {
	return scoreCategorizers[pt].requestScoreCategory(pt, ptInfo, year, friends, players)
}

// Search gets the PlayerSearchResults for the PlayerType/year
func Search(pt db.PlayerType, year int, playerNamePrefix string, activePlayersOnly bool) ([]PlayerSearchResult, error) {
	return searchers[pt].search(pt, year, playerNamePrefix, activePlayersOnly)
}

func (r *httpRequestor) structPointerFromURI(uri string, v interface{}) error {
	b, ok := r.cache.get(uri)
	if !ok {
		var err error
		b, err = r.bytes(uri)
		if err != nil {
			return err
		}
		r.cache.add(uri, b)
	}
	err := json.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("reading json when requesting %v: %w", uri, err)
	}
	return nil
}

func (r *httpRequestor) bytes(uri string) ([]byte, error) {
	if r.logRequestURIs {
		log.Printf("%T : requesting %v", r.httpClient, uri)
	}
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("initializing request to %v: %w", uri, err)
	}
	request.Header.Add("Accept", "application/json")
	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting %v: %w", uri, err)
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body of %v: %w", uri, err)
	}
	return b, nil
}

// ClearCache clears the request cache
func ClearCache() {
	httpCache.clear()
}
