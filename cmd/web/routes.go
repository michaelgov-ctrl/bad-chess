package main

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/michaelgov-ctrl/bad-chess/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /", dynamic.ThenFunc(app.home))

	mux.Handle("GET /matchmaking", dynamic.ThenFunc(app.matchMakingHandler))
	mux.Handle("GET /matches/", dynamic.ThenFunc(app.matchesHandler))
	mux.Handle("GET /matches/ws", dynamic.ThenFunc(app.gameManager.serveWS))

	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
