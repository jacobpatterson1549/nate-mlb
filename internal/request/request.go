package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func requestStruct(url string, v interface{}) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("problem initializing request to %v: %v", url, err)
	}
	request.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("problem requesting %v: %v", url, err)
	}

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
