package main

import (
	"context"
	"flag"
	"html/template"
	"log/slog"
	"os"
	"strings"
)

type config struct {
	port     int
	logLevel string
	limiter  struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config        config
	logger        *slog.Logger
	manager       *Manager
	templateCache map[string]*template.Template
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.logLevel, "log-level", "error", "Logging level (trace|debug|info|warning|error)")

	// TODO: implement rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 20, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 40, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space seperated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(cfg.logLevel)}))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config:        cfg,
		logger:        logger,
		manager:       NewManager(context.Background(), WithLogger(logger)),
		templateCache: templateCache,
	}

	// if err := app.serveTLS("", ""); err != nil {
	if err := app.serve(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
