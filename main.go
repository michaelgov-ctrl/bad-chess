package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	manager := NewManager(context.Background())
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/ping", pingEndpoint)
	mux.HandleFunc("/ws", manager.serveWS)

	log.Println("starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	log.Println(err)
}

func pingEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"copacetic"}`))
}
