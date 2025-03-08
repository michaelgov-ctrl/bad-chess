package main

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Manager interface {
	serveWS(w http.ResponseWriter, r *http.Request)
	addClient(c *Client)
	removeClient(c *Client)
	routeEvent(req Event, c *Client) error
}

type ManagerOptions struct {
	logger   *slog.Logger
	registry *prometheus.Registry
}

type ManagerOption func(*ManagerOptions)

func WithLogger(logger *slog.Logger) ManagerOption {
	return func(m *ManagerOptions) {
		m.logger = logger
	}
}

func WithMetricsRegistry(registry *prometheus.Registry) ManagerOption {
	return func(m *ManagerOptions) {
		m.registry = registry
	}
}
