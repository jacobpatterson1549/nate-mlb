package main

import (
	"fmt"
	"net/http"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	friendPlayerInfo, err := getFriendPlayerInfo()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// jsonOptput := fmt.Sprintf("%v\n", friendPlayerInfo)

	scoreCategory, err := getStats(friendPlayerInfo)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	jsonOptput := fmt.Sprintf("%v\n", scoreCategory)

	// TODO: format to template
	w.Write([]byte(jsonOptput)) // TODO: stream
}

func startServer(portNumber int) {
	http.HandleFunc("/", serveRoot)
	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}
