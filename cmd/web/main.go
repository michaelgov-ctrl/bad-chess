package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/michaelgov-ctrl/bad-chess/internal/models"
	"github.com/michaelgov-ctrl/bad-chess/internal/slogloki"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type config struct {
	port     int
	lokiPort int
	logLevel string
	cert     string
	key      string
	cors     struct {
		trustedOrigins []string
	}
}

type application struct {
	config             config
	authentication     *models.LazyAuth
	engineManager      *EngineManager
	matchmakingManager *MatchmakingManager
	sessionManager     *scs.SessionManager
	templateCache      map[string]*template.Template
	formDecoder        *form.Decoder
	logger             *slog.Logger
	metricsRegistry    *prometheus.Registry
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "API server port")

	flag.IntVar(&cfg.lokiPort, "loki-port", 0, "Port of local loki instace")
	flag.StringVar(&cfg.logLevel, "log-level", "error", "Logging level (trace|debug|info|warning|error)")

	flag.StringVar(&cfg.cert, "cert", "", "File containing cert for tls")
	flag.StringVar(&cfg.key, "key", "", "File containing key for tls")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space seperated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(cfg.logLevel)}))
	if cfg.lokiPort != 0 {
		logger = slogloki.NewLokiLogger("bad-chess", fmt.Sprintf("http://localhost:%d/loki/api/v1/push", cfg.lokiPort), logLevel(cfg.logLevel))
	}

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
		config:             cfg,
		authentication:     models.NewLazyAuth(),
		engineManager:      NewEngineManager(context.Background(), WithLogger(logger), WithMetricsRegistry(registry)),
		matchmakingManager: NewMatchmakingManager(context.Background(), WithLogger(logger), WithMetricsRegistry(registry)),
		sessionManager:     sessionManager,
		templateCache:      templateCache,
		formDecoder:        form.NewDecoder(),
		logger:             logger,
		metricsRegistry:    registry,
	}

	if err := app.serve(cfg.cert, cfg.key); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
