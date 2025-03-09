package game

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// TODO: update this for proxy
	AllowedOrigins = []string{
		"ws://localhost:8080",
		"ws://localhost:8081",
		"ws://localhost:8082",
		"http://localhost:8080",
	}
)

// TODO: update this for proxy
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
