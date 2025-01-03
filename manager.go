package main

import (
	"context"
	"errors"
	"log"
	"net/http"
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
	sync.RWMutex

	clients ClientList

	matches          MatchList
	matchCleanupChan chan MatchId

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:          make(ClientList),
		matches:          make(MatchList),
		matchCleanupChan: make(chan MatchId),
		handlers:         make(map[string]EventHandler),
	}

	m.registerEventHandlers()
	go m.cleanupMatches()

	return m
}

func (m *Manager) registerEventHandlers() {
	m.handlers[EventNewMatchRequest] = m.MatchMakingHandler
	m.handlers[EventMakeMove] = m.MakeMoveHandler
}

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

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection from", r.RemoteAddr)

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)
	//m.matchMaking(client)

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

func (m *Manager) cleanupMatches() {
	cleanupTime, finishedMatches := time.NewTicker(5*time.Second), []MatchId{}
	for {
		select {
		case id, ok := <-m.matchCleanupChan:
			if !ok {
				panic("manager match cleanup channel broken")
			}

			finishedMatches = append(finishedMatches, id)
		case <-cleanupTime.C:
			m.Lock()
			for _, id := range finishedMatches {
				delete(m.matches, id)
			}
			m.Unlock()

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
