package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	AllowedOrigins = []string{
		"ws://localhost:8080",
		"ws://localhost:8081",
		"ws://localhost:8082",
		"http://localhost:8080",
	}
)

type Manager struct {
	logger *slog.Logger

	clients   ClientList
	clientsMu sync.RWMutex

	matches          TimeControlMatchList
	matchesMu        sync.RWMutex
	matchCleanupChan chan MatchOutcome

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		logger:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
		clients: make(ClientList),
		//matches:          make(MatchList),
		matches:          make(TimeControlMatchList),
		matchCleanupChan: make(chan MatchOutcome),
		handlers:         make(map[string]EventHandler),
	}

	m.registerSupportedTimeControls()
	m.registerEventHandlers()
	go m.cleanupMatches()

	return m
}

func (m *Manager) registerEventHandlers() {
	m.handlers[EventJoinMatchRequest] = m.MatchMakingHandler
	m.handlers[EventMakeMove] = m.MakeMoveHandler
}

func (m *Manager) registerSupportedTimeControls() {
	for tc, _ := range SupportedTimeControls {
		m.matches[tc] = make(MatchList)
	}
}

func (m *Manager) addClient(c *Client) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	m.clients[c] = true
}

func (m *Manager) removeClient(c *Client) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	if _, ok := m.clients[c]; ok {
		c.connection.Close()
		delete(m.clients, c)
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
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

func (m *Manager) routeEvent(event Event, c *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return errors.New("there is no such event type")
	}

	if err := handler(event, c); err != nil {
		return err
	}

	return nil
}

func (m *Manager) cleanupMatches() {
	cleanupTime, finishedMatches := time.NewTicker(5*time.Second), []MatchOutcome{}
	for {
		select {
		case matchInfo, ok := <-m.matchCleanupChan:
			if !ok {
				panic("manager match cleanup channel broken")
			}

			finishedMatches = append(finishedMatches, matchInfo)
		case <-cleanupTime.C:
			m.matchesMu.Lock()
			for _, finishedMatch := range finishedMatches {
				m.logger.Info("removing match from manager", "match info", finishedMatch)
				if match, ok := m.matches[finishedMatch.TimeControl][finishedMatch.ID]; ok {
					match.MessagePlayers(Event{Type: EventMatchOver}, Light, Dark)
				}

				delete(m.matches[finishedMatch.TimeControl], finishedMatch.ID)
			}
			m.matchesMu.Unlock()

			finishedMatches = nil
		}
	}
}

func checkOrigin(r *http.Request) bool {
	return true
	/*
		origin := r.Header.Get("Origin")
		for _, o := range AllowedOrigins {
			if origin == o {
				return true
			}
		}

		return false
	*/
}

func (m *Manager) matchMakingAddClientToMatch(c *Client) error {
	match := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]
	outgoingEvent, err := NewOutgoingEvent(EventAssignedMatch, c.currentMatch)
	if err != nil {
		return err
	}

	switch c.currentMatch.Pieces {
	case Light:
		m.matches[c.currentMatch.TimeControl][c.currentMatch.ID].LightPlayer = &Player{Client: c}
		match.MessagePlayers(outgoingEvent, Light)
	case Dark:
		m.matches[c.currentMatch.TimeControl][c.currentMatch.ID].DarkPlayer = &Player{Client: c}
		match.MessagePlayers(outgoingEvent, Dark)
	}

	return nil
}
