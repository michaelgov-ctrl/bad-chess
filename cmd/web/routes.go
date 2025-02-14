package main

import (
	"net/http"

	"github.com/michaelgov-ctrl/badchess/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /ping", app.pingEndpoint)

	mux.HandleFunc("/matches/", app.matchesHandler)
	mux.HandleFunc("/matches/ws", app.manager.serveWS)

	return mux
}
