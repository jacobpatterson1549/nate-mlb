package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func request(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("problem initializing request to %v: %v", url, err)
	}

	request.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	r, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("problem requesting %v: %v", url, err)
	}
	return r, nil
}

func requestStruct(url string, v interface{}) error {
	response, err := request(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&v)
	if err != nil {
		return fmt.Errorf("problem reading json when requesting %v: %v", url, err)
	}
	return nil
}
