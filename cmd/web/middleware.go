package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
	"github.com/prometheus/client_golang/prometheus"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com; img-src 'self' data:; script-src 'self'")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetString(r.Context(), authenticatedSessionKeyName)
		if len(id) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		if app.authentication.Exists(id) {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

// TODO: check here if i have trouble getting setup with a reverse proxy
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)

						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (app *application) metrics(next http.Handler) http.Handler {
	var totalRequestsReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requests_received_total",
			Help: "Total number of http requests received",
		},
	)

	var totalResponsesSent = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "responses_sent_total",
			Help: "Total number of http responses sent",
		},
	)

	var totalResponsesSentByStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "responses_sent_by_status_total",
			Help: "Total http responses sent by status",
		},
		[]string{
			"response_code",
		},
	)

	var requestersNumOfRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_by_requester_total",
			Help: "Total number of requests for each requester",
		},
		[]string{
			"host",
		},
	)

	app.metricsRegistry.MustRegister(totalRequestsReceived, totalResponsesSent, totalResponsesSentByStatus, requestersNumOfRequests)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Inc()

		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		totalResponsesSent.Inc()

		totalResponsesSentByStatus.With(prometheus.Labels{
			"response_code": strconv.Itoa(mw.statusCode),
		}).Inc()

		// TODO: this will have to change to the forwarded for once behind a proxy
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = "parse error"
		}

		requestersNumOfRequests.With(prometheus.Labels{
			"host": ip,
		}).Inc()
	})
}
