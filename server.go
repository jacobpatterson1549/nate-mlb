package main

import (
	"fmt"
	"net/http"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	friendPlayerInfo := getFriendPlayerInfo()
	jsonOptput := fmt.Sprintf("%v", friendPlayerInfo)
	// TODO: get stats on friendPlayerInfo
	// TODO: format to template
	w.Write([]byte(jsonOptput)) // TODO: stream
}

func startServer(portNumber int) {
	http.HandleFunc("/", serveRoot)
	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}
