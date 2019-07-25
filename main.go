package main

import (
	"log"
	"net/http"
	"os"
)

func serveRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!\n"))
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port not set.")
	}

	http.HandleFunc("/", serveRoot)
	http.ListenAndServe(":"+port, nil)
}
