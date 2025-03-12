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
type MatchmakingManager struct {
	clients   ClientList
	clientsMu sync.RWMutex

	matches          TimeControlMatchList
	matchesMu        sync.RWMutex
	matchCleanupChan chan MatchOutcome

	handlers map[string]EventHandler

	ManagerOptions
	metrics *MatchmakingManagerMetrics
}

func NewMatchmakingManager(ctx context.Context, opts ...ManagerOption) *MatchmakingManager {
	m := &MatchmakingManager{
		clients:          make(ClientList),
		matches:          make(TimeControlMatchList),
		matchCleanupChan: make(chan MatchOutcome),
		handlers:         make(map[string]EventHandler),
		metrics:          &MatchmakingManagerMetrics{},
	}

	defaults := &ManagerOptions{
		logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
		registry: prometheus.NewRegistry(),
	}

	for _, opt := range opts {
		opt(defaults)
	}

	m.ManagerOptions = *defaults

	m.registerMatchmakingManagerMetrics()
	m.registerSupportedTimeControls()
	m.registerEventHandlers()
	go m.cleanupMatches()

	return m
}

// TODO: add a handler for joining with a valid match id
func (m *MatchmakingManager) registerEventHandlers() {
	m.handlers[EventJoinMatchRequest] = m.matchMakingHandler
	m.handlers[EventMakeMove] = m.makeMoveHandler
}

func (m *MatchmakingManager) registerSupportedTimeControls() {
	m.matchesMu.Lock()
	defer m.matchesMu.Unlock()

	for tc := range SupportedTimeControls {
		m.matches[tc] = make(MatchList)
	}
}

func (m *MatchmakingManager) addClient(c *Client) {
	m.metrics.totalClients.Inc()
	m.logger.Debug("new client", "client", c)

	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	m.clients[c] = true
}

func (m *MatchmakingManager) removeClient(c *Client) {
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

func (m *MatchmakingManager) ServeWS(w http.ResponseWriter, r *http.Request) {
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

func (m *MatchmakingManager) routeEvent(event Event, c *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return errors.New("there is no such event type")
	}

	if err := handler(event, c); err != nil {
		return err
	}

	return nil
}

// From the context of games coming from the website it makes sense to close client connections here
// TODO: it should be more graceful
func (m *MatchmakingManager) cleanupMatches() {
	cleanupTime, finishedMatches := time.NewTicker(5*time.Second), []MatchOutcome{}
	for {
		select {
		case matchInfo, ok := <-m.matchCleanupChan:
			if !ok {
				panic("Matchmaking Manager match cleanup channel broken")
			}

			finishedMatches = append(finishedMatches, matchInfo)
		case <-cleanupTime.C:
			m.matchesMu.Lock()
			for _, finishedMatch := range finishedMatches {
				m.logger.Debug("removing match from Matchmaking Manager", "match info", finishedMatch)

				if match, ok := m.matches[finishedMatch.TimeControl][finishedMatch.ID]; ok {
					match.MessagePlayers(Event{Type: EventMatchOver}, Light, Dark)

					if match.LightPlayer != nil {
						m.removeClient(match.LightPlayer.Client)
					}

					if match.DarkPlayer != nil {
						m.removeClient(match.DarkPlayer.Client)
					}
				}

				delete(m.matches[finishedMatch.TimeControl], finishedMatch.ID)
			}
			m.matchesMu.Unlock()

			finishedMatches = nil
		}
	}
}

func (m *MatchmakingManager) addClientToMatch(c *Client) error {
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

func (m *MatchmakingManager) matchMakingHandler(event Event, c *Client) error {
	m.logger.Info("match making handler", "event", event, "client", *c)

	var joinEvent JoinMatchEvent
	if err := json.Unmarshal(event.Payload, &joinEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	if _, ok := SupportedTimeControls[joinEvent.TimeControl]; !ok {
		return fmt.Errorf("unsupported time control")
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
			err := m.addClientToMatch(c)
			if err != nil {
				return err
			}

			// both players should now be present to start game
			return match.Start(m.matchCleanupChan)
		}
	}

	c.currentMatch.ID = m.newMatch(joinEvent.TimeControl)
	c.currentMatch.Pieces = Light

	err := m.addClientToMatch(c)
	if err == nil {
		m.metrics.totalMatches.Inc()
	}

	return err
}

func (m *MatchmakingManager) makeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	m.matchesMu.RLock()
	defer m.matchesMu.RUnlock()

	match, ok := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]
	if !ok {
		return errors.New("no match")
	}

	clientPlayerColor := match.ClientPieceColor(c)

	if !match.OpponentPresent(clientPlayerColor) {
		return fmt.Errorf("no opponent present")
	}

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

	return nil
}

// TODO: this is probably fine for now, consider tossing an error on collision & retrying
func (m *MatchmakingManager) newMatch(timeControl TimeControl) MatchId {
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
			Logger:      m.logger,
		}

		go match.notifyIfStale(m.matchCleanupChan)

		m.matches[timeControl][matchId] = match
	}

	return matchId
}

type MatchmakingManagerMetrics struct {
	totalClients   prometheus.Counter
	currentClients prometheus.Gauge
	totalMatches   prometheus.Counter
	currentMatches prometheus.Gauge
}

func (m *MatchmakingManager) registerMatchmakingManagerMetrics() {
	m.metrics.totalClients = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "matchmaking_manager_clients_total",
			Help: "Total number of clients the matchmaking manager has handled",
		},
	)

	m.metrics.currentClients = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "matchmaking_manager_clients_current",
			Help: "Current number of connected matchmaking clients",
		},
	)

	m.metrics.totalMatches = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "matchmaking_manager_matches_total",
			Help: "Total number of matches the matchmaking manager has handled",
		},
	)

	m.metrics.currentMatches = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "matchmaking_manager_matches_current",
			Help: "Current number of matchmaking matches",
		},
	)

	m.registry.MustRegister(m.metrics.totalClients, m.metrics.currentClients, m.metrics.totalMatches, m.metrics.currentMatches)

	go m.updateMetrics()
}

func (m *MatchmakingManager) updateMetrics() {
	for {
		time.Sleep(5 * time.Second)
		go m.updateCurrentClientsMetric()
		go m.updateCurrentMatchesMetric()
	}
}

func (m *MatchmakingManager) updateCurrentClientsMetric() {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	m.metrics.currentClients.Set(float64(len(m.clients)))
}

func (m *MatchmakingManager) updateCurrentMatchesMetric() {
	m.matchesMu.RLock()
	defer m.matchesMu.RUnlock()

	var sum int
	for _, matchList := range m.matches {
		sum += len(matchList)
	}

	m.metrics.currentMatches.Set(float64(sum))
}
