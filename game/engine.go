package game

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
	"github.com/notnil/chess"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: Handle unsupported time controls & engine ELOs by returning error
type EngineManager struct {
	clients   ClientList
	clientsMu sync.RWMutex

	matches          ELOMatchList
	matchesMu        sync.RWMutex
	matchCleanupChan chan EngineMatchOutcome

	handlers map[string]EventHandler

	ManagerOptions
	metrics *EngineManagerMetrics
}

func NewEngineManager(ctx context.Context, opts ...ManagerOption) *EngineManager {
	m := &EngineManager{
		clients:          make(ClientList),
		matches:          make(ELOMatchList),
		matchCleanupChan: make(chan EngineMatchOutcome),
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

	go m.cleanupMatches()
	/*
		go m.updateMetrics()
	*/

	return m
}

func (m *EngineManager) registerEventHandlers() {
	m.handlers[EventNewEngineMatchRequest] = m.engineMatchRequestHandler
	m.handlers[EventMakeMove] = m.makeMoveHandler
}

func (m *EngineManager) registerSupportedEngineELOs() {
	m.matchesMu.Lock()
	defer m.matchesMu.Unlock()

	for ee := range SupportedEngineELOs {
		m.matches[ee] = make(EngineMatchList)
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

func (m *EngineManager) engineMatchRequestHandler(event Event, c *Client) error {
	m.logger.Info("match making handler", "event", event, "client", *c)

	newMatchEvent, err := m.parseMatchRequest(event)
	if err != nil {
		return err
	}

	m.matchesMu.RLock()
	matchId := m.newMatchId(newMatchEvent.ELO)
	m.matchesMu.RUnlock()

	playerPieces := assignPlayerPieces()

	match, err := m.newEngineMatch(matchId, newMatchEvent.ELO, c, playerPieces)
	if err != nil {
		return err
	}

	m.matchesMu.Lock()
	m.matches[newMatchEvent.ELO][matchId] = match
	m.matchesMu.Unlock()

	c.currentMatch = NewClientMatchInfo(matchId, Engine, EngineMatchTimeControl, newMatchEvent.ELO, playerPieces)
	outgoingEvent, err := NewOutgoingEvent(EventAssignedMatch, c.currentMatch)
	if err != nil {
		return err
	}

	match.messagePlayer(outgoingEvent)

	if err := match.Start(m.matchCleanupChan); err != nil {
		return err
	}

	if playerPieces != Light {
		return match.EngineMove()
	}

	return nil
}

func (m *EngineManager) parseMatchRequest(event Event) (NewEngineMatchEvent, error) {
	var newMatchEvent NewEngineMatchEvent
	if err := json.Unmarshal(event.Payload, &newMatchEvent); err != nil {
		return NewEngineMatchEvent{}, fmt.Errorf("bad payload in request: %w", err)
	}

	if _, ok := SupportedEngineELOs[newMatchEvent.ELO]; !ok {
		return NewEngineMatchEvent{}, fmt.Errorf("unsupported engine ELO: %d", newMatchEvent.ELO)
	}

	return newMatchEvent, nil
}

func (m *EngineManager) newEngineMatch(matchId MatchId, elo ELO, c *Client, playerPieces PieceColor) (*EngineMatch, error) {
	engine, err := NewEngine(elo)
	if err != nil {
		return nil, err
	}

	match := &EngineMatch{
		ID:     matchId,
		Engine: engine,
		Player: &Player{
			Client: c,
			Clock:  NewClock(EngineMatchTimeControl),
		},
		PlayerPieces: playerPieces,
		Game:         chess.NewGame(),
		Turn:         Light,
		State:        Waiting,
		Logger:       m.logger,
	}
	go match.notifyIfStale(m.matchCleanupChan)

	return match, nil
}

func (m *EngineManager) newMatchId(elo ELO) MatchId {
	matchId := MatchId(uuid.NewString())

	if _, ok := m.matches[elo][matchId]; ok {
		m.logger.Error("uuid collision", "MatchId", matchId)
		matchId = m.newMatchId(elo)
	}

	return matchId
}

func (m *EngineManager) makeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	m.matchesMu.RLock()
	defer m.matchesMu.RUnlock()

	match, ok := m.matches[c.currentMatch.EngineELO][c.currentMatch.ID]
	if !ok {
		return errors.New("no match")
	}

	if err := match.MakeMove(c.currentMatch.Pieces, moveEvent.Move); err != nil {
		return err
	}

	match.Player.Clock.Pause()
	defer match.Player.Clock.Start()
	// if engine move errors player clock will be started
	// which is fine for now since the game state will be ruined if the engine move fails
	// TODO: handle that

	return match.EngineMove()
}

// From the context of games coming from the website it makes sense to close client connections here
// TODO: it should be more graceful
func (m *EngineManager) cleanupMatches() {
	cleanupTime, finishedMatches := time.NewTicker(5*time.Second), []EngineMatchOutcome{}
	for {
		select {
		case matchInfo, ok := <-m.matchCleanupChan:
			if !ok {
				panic("Engine Manager match cleanup channel broken")
			}

			finishedMatches = append(finishedMatches, matchInfo)
		case <-cleanupTime.C:
			m.matchesMu.Lock()
			for _, finishedMatch := range finishedMatches {
				m.logger.Debug("removing match from Engine Manager", "match info", finishedMatch)

				if match, ok := m.matches[finishedMatch.ELO][finishedMatch.ID]; ok {
					match.messagePlayer(Event{Type: EventMatchOver})

					if match.Player != nil {
						m.removeClient(match.Player.Client)
					}
				}

				delete(m.matches[finishedMatch.ELO], finishedMatch.ID)
			}
			m.matchesMu.Unlock()

			finishedMatches = nil
		}
	}
}

type EngineManagerMetrics struct {
}

func NewEngineManagerMetrics() *EngineManagerMetrics {
	return &EngineManagerMetrics{}
}
