package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
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

// TODO: evaluate passing the application logger as the Manager's logger
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

// TODO: add a handler for joining with a valid match id
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

func (m *Manager) MatchMakingHandler(event Event, c *Client) error {
	m.logger.Info("match making handler", "event", event, "client", *c)

	var joinEvent JoinMatchEvent
	if err := json.Unmarshal(event.Payload, &joinEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	c.currentMatch.TimeControl = joinEvent.TimeControl

	// this is probably slow
	m.matchesMu.Lock()
	defer m.matchesMu.Unlock()
	for matchId, match := range m.matches[joinEvent.TimeControl] {
		// no created match should ever be missing a light player since theyre added at creation
		if match.DarkPlayer == nil {
			m.matches[joinEvent.TimeControl][matchId].DarkPlayer = &Player{Client: c}
			c.currentMatch.ID = matchId
			c.currentMatch.Pieces = Dark
			err := m.matchMakingAddClientToMatch(c)
			if err != nil {
				return err
			}

			// both players should now be present to start game
			return match.Start(m.matchCleanupChan)
		}
	}

	c.currentMatch.ID = m.newMatch(joinEvent.TimeControl)
	c.currentMatch.Pieces = Light
	err := m.matchMakingAddClientToMatch(c)
	return err
}

func (m *Manager) MakeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	m.matchesMu.RLock()
	defer m.matchesMu.RUnlock()
	if match, ok := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]; ok {
		clientPlayerColor := match.ClientPieceColor(c)
		if clientPlayerColor == NoColor || clientPlayerColor != c.currentMatch.Pieces {
			return fmt.Errorf("player pieces are borked")
		}

		if err := match.MakeMove(c.currentMatch.Pieces, moveEvent.Move); err != nil {
			return err
		}

		outgoingEvent, err := NewOutgoingEvent(EventPropagatePosition, PropagatePositionEvent{
			PlayerColor: clientPlayerColor.String(),
			FEN:         match.Game.FEN(),
		})
		if err != nil {
			return err
		}

		// egress is handled in (c *Client) writeMessages()
		match.MessagePlayers(outgoingEvent, Dark, Light)
	} else {
		return errors.New("no match")
	}

	return nil
}

// TODO: this is probably fine for now, consider tossing an error on collision & retrying
func (m *Manager) newMatch(timeControl TimeControl) MatchId {
	matchId := MatchId(uuid.NewString())

	// right now newMatch() is only called in matchmaking which locks m
	if _, ok := m.matches[timeControl][matchId]; ok {
		m.logger.Error("uuid collision", "MatchId", matchId)
		matchId = m.newMatch(timeControl)
	} else {
		match := &Match{
			ID:          matchId,
			TimeControl: timeControl,
			Game:        chess.NewGame(),
			Turn:        Light,
			State:       Waiting,
		}

		go match.notifyIfStale(m.matchCleanupChan)

		m.matches[timeControl][matchId] = match
	}

	return matchId
}
