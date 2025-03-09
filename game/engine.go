package game

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type EngineManager struct {
	clients   ClientList
	clientsMu sync.RWMutex

	matches          ELOMatchList
	matchesMu        sync.RWMutex
	matchCleanupChan chan MatchOutcome

	handlers map[string]EventHandler

	ManagerOptions
	metrics *EngineManagerMetrics
}

func NewEngineManager(ctx context.Context, opts ...ManagerOption) *EngineManager {
	m := &EngineManager{
		clients:          make(ClientList),
		matches:          make(ELOMatchList),
		matchCleanupChan: make(chan MatchOutcome),
		handlers:         make(map[string]EventHandler),
		metrics:          NewEngineManagerMetrics(),
	}

	defaults := &ManagerOptions{
		logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
		registry: prometheus.NewRegistry(),
	}

	for _, opt := range opts {
		opt(defaults)
	}

	m.ManagerOptions = *defaults

	/*
		metrics go here
	*/

	m.registerSupportedEngineELOs()
	m.registerEventHandlers()

	/*
		go m.cleanupMatches()
		go m.updateMetrics()
	*/

	return m
}

func (m *EngineManager) registerEventHandlers() {
	/*
		m.handlers[EventJoinMatchRequest] = m.EngineHandler
		m.handlers[EventMakeMove] = m.MakeMoveHandler
	*/
}

func (m *EngineManager) registerSupportedEngineELOs() {
	m.matchesMu.Lock()
	defer m.matchesMu.Unlock()

	for ee := range SupportedEngineELOs {
		m.matches[ee] = make(MatchList)
	}
}

func (m *EngineManager) addClient(c *Client) {
	m.logger.Debug("new client", "client", c)

	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	m.clients[c] = true
}

func (m *EngineManager) removeClient(c *Client) {
	m.logger.Debug("removed client", "client", c)

	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	if _, ok := m.clients[c]; ok {
		if c != nil {
			c.connection.Close()
		}
		delete(m.clients, c)
	}
}

func (m *EngineManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	m.logger.Info("new connection", "origin", r.RemoteAddr)

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error(err.Error())
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)

	go client.readEvents(m.logger)
	go client.writeEvents(m.logger)
}

func (m *EngineManager) routeEvent(event Event, c *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return errors.New("there is no such event type")
	}

	if err := handler(event, c); err != nil {
		return err
	}

	return nil
}

type EngineManagerMetrics struct {
}

func NewEngineManagerMetrics() *EngineManagerMetrics {
	return &EngineManagerMetrics{}
}
