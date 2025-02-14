package main

import (
	"net/http"

	"github.com/michaelgov-ctrl/badchess/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /", http.FileServerFS(ui.Files))

	mux.HandleFunc("/ping", pingEndpoint)
	mux.HandleFunc("/ws", app.manager.serveWS)

	return mux
}
