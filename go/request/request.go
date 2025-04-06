// Package request requests scores and performs searches for different sports/players.
package request

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

type (
	requester interface {
		structPointerFromURI(uri string, v interface{}) error
	}

	// HTTPClient makes HTTP requests
	HTTPClient interface {
		Do(r *http.Request) (*http.Response, error)
	}

	httpRequester struct {
		cache          Cache
		httpClient     HTTPClient
		logRequestURIs bool
		log            *log.Logger
	}
)

// NewRequesters creates new ScoreCategorizers and Searchers for the specified PlayerTypes and an aboutRequester
func NewRequesters(httpClient HTTPClient, c Cache, nflAppKey, environment string, logRequestURIs bool, log *log.Logger) (map[db.PlayerType]ScoreCategorizer, map[db.PlayerType]Searcher, AboutRequester) {
	r := httpRequester{
		cache:          c,
		httpClient:     httpClient,
		logRequestURIs: logRequestURIs,
		log:            log,
	}
	nflR := nflRequester{
		appKey:    nflAppKey,
		requester: &r,
	}

	mlbTeamR := mlbTeamRequester{requester: &r}
	mlbPlayerR := mlbPlayerRequester{requester: &r}
	mlbPlayerS := mlbPlayerSearcher{requester: &r}
	nflTeamR := nflTeamRequester{requester: &nflR}
	nflPlayerR := nflPlayerRequester{requester: &nflR}

	scoreCategorizers := make(map[db.PlayerType]ScoreCategorizer)
	scoreCategorizers[db.PlayerTypeMlbTeam] = &mlbTeamR
	scoreCategorizers[db.PlayerTypeMlbHitter] = &mlbPlayerR
	scoreCategorizers[db.PlayerTypeMlbPitcher] = &mlbPlayerR
	scoreCategorizers[db.PlayerTypeNflTeam] = &nflTeamR
	scoreCategorizers[db.PlayerTypeNflQB] = &nflPlayerR
	scoreCategorizers[db.PlayerTypeNflMisc] = &nflPlayerR

	searchers := make(map[db.PlayerType]Searcher)
	searchers[db.PlayerTypeMlbTeam] = &mlbTeamR
	searchers[db.PlayerTypeMlbHitter] = &mlbPlayerS
	searchers[db.PlayerTypeMlbPitcher] = &mlbPlayerS
	searchers[db.PlayerTypeNflTeam] = &nflTeamR
	searchers[db.PlayerTypeNflQB] = &nflPlayerR
	searchers[db.PlayerTypeNflMisc] = &nflPlayerR

	aboutRequester := AboutRequester{environment: environment, requester: &r}

	return scoreCategorizers, searchers, aboutRequester
}

func (r *httpRequester) structPointerFromURI(uri string, v interface{}) error {
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

func (r *httpRequester) bytes(uri string) ([]byte, error) {
	if r.logRequestURIs {
		r.log.Printf("%T : requesting %v", r.httpClient, uri)
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
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected ok response, got %v", response.StatusCode)
	}

	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body of %v: %w", uri, err)
	}
	return b, nil
}
