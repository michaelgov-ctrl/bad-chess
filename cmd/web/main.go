package main

import (
	"context"
	"flag"
	"html/template"
	"log/slog"
	"os"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/michaelgov-ctrl/bad-chess/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type config struct {
	port     int
	logLevel string
	cors     struct {
		trustedOrigins []string
	}
}

type application struct {
	config          config
	authentication  *models.LazyAuth
	gameManager     *Manager
	sessionManager  *scs.SessionManager
	templateCache   map[string]*template.Template
	formDecoder     *form.Decoder
	logger          *slog.Logger
	metricsRegistry *prometheus.Registry
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.logLevel, "log-level", "error", "Logging level (trace|debug|info|warning|error)")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space seperated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	var cert, key string
	flag.StringVar(&cert, "cert", "server.crt", "File containing cert for tls")
	flag.StringVar(&key, "key", "server.key", "File containing key for tls")

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(cfg.logLevel)}))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = models.MaxSessionAge
	sessionManager.Cookie.Secure = true

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	app := &application{
		config:          cfg,
		authentication:  models.NewLazyAuth(),
		gameManager:     NewManager(context.Background(), WithLogger(logger), WithMetricsRegistry(registry)),
		sessionManager:  sessionManager,
		templateCache:   templateCache,
		formDecoder:     form.NewDecoder(),
		logger:          logger,
		metricsRegistry: registry,
	}

	//if err := app.serveTLS(cert, key); err != nil {
	if err := app.serve(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
