package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type requestor struct {
	cache          cache
	httpClient     *http.Client
	logRequestUrls bool
}

var request requestor

func init() {
	request = requestor{
		cache: newCache(100),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logRequestUrls: false,
	}
}

func (r *requestor) structPointerFromURL(url string, v interface{}) error {
	var (
		b   []byte
		err error
		ok  bool
	)
	if b, ok = request.cache.get(url); !ok {
		b, err = request.bytes(url)
		if err != nil {
			return err
		}
		request.cache.add(url, b)
	}
	json.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("problem reading json when requesting %v: %v", url, err)
	}
	return nil
}

func (r *requestor) bytes(url string) ([]byte, error) {
	if r.logRequestUrls {
		log.Println("Requesting", url)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("problem initializing request to %v: %v", url, err)
	}
	request.Header.Add("Accept", "application/json")
	response, err := r.httpClient.Do(request)
	if err != nil {
		if r.logRequestUrls {
			log.Panicln(" -> FAILED")
		}
		return nil, fmt.Errorf("problem requesting %v: %v", url, err)
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("problem reading body of %v", url)
	}
	return b, nil
}

// ClearCache clears the request cache
func ClearCache() {
	request.cache.clear()
}
