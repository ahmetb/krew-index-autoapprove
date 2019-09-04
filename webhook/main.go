package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT env var is not set")
	}
	// TODO: detect TOKEN is set

	http.HandleFunc("/webhook", webhook)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func webhook(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: detect event type, allow only synchronize or create
	// TODO: find if patch is bump
	// TODO: comment the result back on github PR
}
