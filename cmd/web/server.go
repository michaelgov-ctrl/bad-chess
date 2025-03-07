package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve(certFile, keyFile string) error {
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", app.config.port),
		MaxHeaderBytes: (1024 * 1024) / 2, // half MB
		Handler:        app.routes(),
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		ErrorLog:       slog.NewLogLogger(app.logger.Handler(), logLevel(app.config.logLevel)),
	}

	if certFile != "" && keyFile != "" {
		srv.TLSConfig = &tls.Config{CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256}}
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr)

	err := upgrade(srv, certFile, keyFile)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

func upgrade(srv *http.Server, certFile, keyFile string) error {
	if certFile != "" && keyFile != "" {
		return srv.ListenAndServeTLS(certFile, keyFile)
	}

	return srv.ListenAndServe()
}

func logLevel(s string) slog.Level {
	m := map[string]slog.Level{
		"trace":   slog.Level(-8),
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
	}

	if level, ok := m[s]; ok {
		return level
	}

	return slog.LevelError
}
