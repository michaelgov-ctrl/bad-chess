package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/michaelgov-ctrl/bad-chess/ui"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServerFS(ui.Files))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))

	router.Handler(http.MethodGet, "/engineselection", protected.ThenFunc(app.engineSelectionHandler))
	router.Handler(http.MethodGet, "/engines", protected.ThenFunc(app.enginesHandler))
	router.Handler(http.MethodGet, "/engines/ws", protected.ThenFunc(app.engineManager.ServeWS))

	router.Handler(http.MethodGet, "/matchmaking", protected.ThenFunc(app.matchMakingHandler))
	router.Handler(http.MethodGet, "/matches", protected.ThenFunc(app.matchesHandler))
	router.Handler(http.MethodGet, "/matches/ws", protected.ThenFunc(app.matchmakingManager.ServeWS))

	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	router.Handler(http.MethodGet, "/metrics", promhttp.HandlerFor(app.metricsRegistry, promhttp.HandlerOpts{}))

	standard := alice.New(app.metrics, app.recoverPanic, app.enableCORS, app.logRequest, secureHeaders)

	return standard.Then(router)
}
