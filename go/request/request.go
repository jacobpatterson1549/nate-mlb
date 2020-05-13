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
	requester interface {
		structPointerFromURI(uri string, v interface{}) error
	}

	httpClient interface {
		Do(r *http.Request) (*http.Response, error)
	}

	httpRequester struct {
		cache          Cache
		httpClient     httpClient
		logRequestURIs bool
		log            *log.Logger
	}
)

// NewRequesters creates new ScoreCategorizers and Searchers for the specified PlayerTypes and an aboutRequester
func NewRequesters(c Cache, nflAppKey, environment string, log *log.Logger) (map[db.PlayerType]ScoreCategorizer, map[db.PlayerType]Searcher, AboutRequester) {
	r := httpRequester{
		cache: c,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logRequestURIs: false,
		log:            log,
	}
	nflRequester := nflRequester{
		appKey:    nflAppKey,
		requester: &r,
	}

	mlbTeamRequester := mlbTeamRequester{requester: &r}
	mlbPlayerScoreCategorizer := mlbPlayerRequester{requester: &r}
	mlbPlayerSearcher := mlbPlayerSearcher{requester: &r}
	nflTeamRequester := nflTeamRequester{requester: &nflRequester}
	nflPlayerRequester := nflPlayerRequester{requester: &nflRequester}

	scoreCategorizers := make(map[db.PlayerType]ScoreCategorizer)
	scoreCategorizers[db.PlayerTypeMlbTeam] = &mlbTeamRequester
	scoreCategorizers[db.PlayerTypeMlbHitter] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeMlbPitcher] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeNflTeam] = &nflTeamRequester
	scoreCategorizers[db.PlayerTypeNflQB] = &nflPlayerRequester
	scoreCategorizers[db.PlayerTypeNflMisc] = &nflPlayerRequester

	searchers := make(map[db.PlayerType]Searcher)
	searchers[db.PlayerTypeMlbTeam] = &mlbTeamRequester
	searchers[db.PlayerTypeMlbHitter] = &mlbPlayerSearcher
	searchers[db.PlayerTypeMlbPitcher] = &mlbPlayerSearcher
	searchers[db.PlayerTypeNflTeam] = &nflTeamRequester
	searchers[db.PlayerTypeNflQB] = &nflPlayerRequester
	searchers[db.PlayerTypeNflMisc] = &nflPlayerRequester

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
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body of %v: %w", uri, err)
	}
	return b, nil
}
