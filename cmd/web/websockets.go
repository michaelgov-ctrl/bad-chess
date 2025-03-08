package main

import "github.com/gorilla/websocket"

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
