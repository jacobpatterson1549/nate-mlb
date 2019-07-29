package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	friendPlayerInfo, err := getFriendPlayerInfo()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	scoreCategories, err := getStats(friendPlayerInfo)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonOptput, err := json.Marshal(scoreCategories)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(jsonOptput)
}

func startServer(portNumber int) {
	http.HandleFunc("/", serveRoot)
	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}
