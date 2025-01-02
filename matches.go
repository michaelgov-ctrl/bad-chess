package main

import (
	"errors"
	"log"
	"time"

	"github.com/notnil/chess"
)

type MatchId string

// TODO: Spectators  []*Client
type Match struct {
	ID MatchId

	LightPlayer *Client
	DarkPlayer  *Client

	*ChessGame
}

// TODO: cleanup matches
type MatchList map[MatchId]*Match

func (m *Match) ClientPlayerColor(client *Client) (PlayerColor, error) {
	if m.LightPlayer != nil && m.LightPlayer == client {
		return Light, nil
	}

	if m.DarkPlayer != nil && m.DarkPlayer == client {
		return Dark, nil
	}

	return NoColor, errors.New("missing player")
}

func (m *Match) notifyWhenOver(ch chan<- MatchId) {
	started, waitTime := time.Now(), 30*time.Second
	for m.Game.Outcome() == chess.NoOutcome && time.Since(started) <= waitTime {
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("signaling to close", m.ID)
	ch <- m.ID
}
