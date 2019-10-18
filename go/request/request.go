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
		cache          Cache
		httpClient     httpClient
		logRequestURIs bool
	}
)

// NewRequestors creates new ScoreCategorizers and Searchers for the specified PlayerTypes and an aboutRequestor
// TODO: rename requestor to requester
func NewRequestors(c Cache) (map[db.PlayerType]ScoreCategorizer, map[db.PlayerType]Searcher, AboutRequestor) {
	r := httpRequestor{
		cache: c,
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

	scoreCategorizers := make(map[db.PlayerType]ScoreCategorizer)
	scoreCategorizers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	scoreCategorizers[db.PlayerTypeMlbHitter] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeMlbPitcher] = &mlbPlayerScoreCategorizer
	scoreCategorizers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	scoreCategorizers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	scoreCategorizers[db.PlayerTypeNflMisc] = &nflPlayerRequestor

	searchers := make(map[db.PlayerType]Searcher)
	searchers[db.PlayerTypeMlbTeam] = &mlbTeamRequestor
	searchers[db.PlayerTypeMlbHitter] = &mlbPlayerSearcher
	searchers[db.PlayerTypeMlbPitcher] = &mlbPlayerSearcher
	searchers[db.PlayerTypeNflTeam] = &nflTeamRequestor
	searchers[db.PlayerTypeNflQB] = &nflPlayerRequestor
	searchers[db.PlayerTypeNflMisc] = &nflPlayerRequestor

	aboutRequestor := AboutRequestor{requestor: &r}

	return scoreCategorizers, searchers, aboutRequestor
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
