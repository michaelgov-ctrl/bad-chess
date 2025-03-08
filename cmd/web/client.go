package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10 // 90% of pongWait
)

type Client struct {
	connection *websocket.Conn
	manager    Manager

	currentMatch ClientMatchInfo

	// egress is used to avoid concurrent writes on the websocket connection for events
	egress chan Event
}

type ClientList map[*Client]bool

type ClientMatchInfo struct {
	ID          MatchId     `json:"match_id"`
	TimeControl TimeControl `json:"time_control"`
	Pieces      PieceColor  `json:"pieces"`
}

func NewClient(conn *websocket.Conn, manager Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readEvents(logger *slog.Logger) {
	defer func() {
		c.manager.removeClient(c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logger.Error(err.Error())
		return
	}

	// prevent maliciously large messages, limited to 512 bytes
	c.connection.SetReadLimit(512)
	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("error reading message", "error", err)
			}
			break
		}

		var req Event
		logger.Debug("received payload", "payload", string(payload))
		if err := json.Unmarshal(payload, &req); err != nil {
			logger.Error("error marshalling event", "error", err)
			break
		}

		if err := c.manager.routeEvent(req, c); err != nil {
			logger.Error("error handling message", "error", err)
			// TODO: switch on error types or otherwise handle them
			c.egress <- Event{
				Payload: []byte(fmt.Sprintf(`{"error":"%v"}`, err)),
				Type:    EventMatchError,
			}
		}
	}
}

func (c *Client) writeEvents(logger *slog.Logger) {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		// bottle necking to prevent abuse of concurrency from client
		select {
		case message, ok := <-c.egress:
			if !ok {
				// if egress is broken notify client & close
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					logger.Error("connection closed", "error", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				logger.Error("error marshalling message", "error", err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				logger.Error("failed to send message", "error", err)
				return
			}

			logger.Debug("message sent")
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				logger.Error("ping error", "error", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(pongMsg string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
