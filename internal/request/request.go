package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var requestCache = newCache(100)

func requestStruct(url string, v interface{}) error {
	var (
		b   []byte
		err error
		ok  bool
	)
	if b, ok = requestCache.get(url); !ok {
		b, err = requestBytes(url)
		if err != nil {
			return err
		}
		requestCache.add(url, b)
	}
	json.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("problem reading json when requesting %v: %v", url, err)
	}
	return nil
}

func requestBytes(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("problem initializing request to %v: %v", url, err)
	}
	request.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := client.Do(request)
	if err != nil {
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
	requestCache.clear()
}
