package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	friendPlayerInfo, err := getFriendPlayerInfo()
	if err == nil {
		scoreCategories, err := getStats(friendPlayerInfo)
		if err == nil {
			jsonOptput, err := json.Marshal(scoreCategories)
			if err == nil {
				w.Write(jsonOptput)
			}
		}
	}

	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func startServer(portNumber int) {
	http.HandleFunc("/", serveRoot)
	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}
