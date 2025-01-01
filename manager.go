package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
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
	sync.RWMutex

	clients ClientList
	matches MatchList

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		matches:  make(MatchList),
		handlers: make(map[string]EventHandler),
	}

	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventMakeMove] = MakeMoveHandler
}

func MakeMoveHandler(event Event, c *Client) error {
	fmt.Printf("event: %v\nclient: %v\n", event, *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	//validate move & move in game as client
	// *here*
	//validate move & move in game as client

	propEvent := PropagateMoveEvent{
		moveEvent,
	}

	data, err := json.Marshal(propEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal propagation event: %v", err)
	}

	outgoingEvent := Event{
		Payload: data,
		Type:    EventPropagateMove,
	}

	c.manager.Lock()
	defer c.manager.Unlock()
	if match, ok := c.manager.matches[c.match]; ok {
		clientPlayerColor, err := match.ClientPlayerColor(c)
		if err != nil {
			c.egress <- Event{
				// TODO: payload should probably maybe be ErrorEvent
				Payload: []byte(`{error:"no color assigned for this match"}`),
				Type:    EventMatchError,
			}
			return err
		}

		// send only to other player, should an accept be sent back to client?
		// egress is handled in (c *Client) writeMessages()
		switch clientPlayerColor {
		case Light:
			if match.DarkPlayer != nil {
				match.DarkPlayer.egress <- outgoingEvent
			}
		case Dark:
			if match.LightPlayer != nil {
				match.LightPlayer.egress <- outgoingEvent
			}
		}
	} else {
		c.egress <- Event{
			// TODO: payload should probably maybe be ErrorEvent
			Payload: []byte(`{error:"client not assigned a match"}`),
			Type:    EventMatchError,
		}
		return errors.New("no match")
	}

	return nil
}

/*
func (m *Manager) PropagateMoveHandler(event Event, c *Client) error {
	fmt.Printf("event: %v\nclient: %v\n", event, *c)
	if match, ok := m.matches[c.match]; ok {
		fmt.Printf("match: %v\n", *match)
	}
	return nil
}
*/

func (m *Manager) addClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[c] = true
}

func (m *Manager) removeClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[c]; ok {
		c.connection.Close()
		delete(m.clients, c)
	}
}

// TODO: this is probably fine for now, consider tossing an error on collision & retrying
func (m *Manager) newMatch() MatchId {
	matchId := MatchId(uuid.NewString())

	// right now newMatch() is only called in matchmaking which locks m
	if _, ok := m.matches[matchId]; ok {
		log.Println("uuid collision", matchId)
		matchId = m.newMatch()
	} else {
		m.matches[matchId] = &Match{Game: newChessGame()}
	}

	return matchId
}

func (m *Manager) matchMaking(c *Client) {
	// this is probably slow
	m.Lock()
	defer m.Unlock()
	for matchId, match := range m.matches {
		if match.LightPlayer == nil {
			m.matches[matchId].LightPlayer = c
			c.match = matchId
			return
		}

		if match.DarkPlayer == nil {
			m.matches[matchId].DarkPlayer = c
			c.match = matchId
			return
		}
	}

	matchId := m.newMatch()
	m.matches[matchId].LightPlayer = c
	c.match = matchId
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection from", r.RemoteAddr)

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)
	m.matchMaking(client)

	go client.readEvents()
	go client.writeEvents()
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
