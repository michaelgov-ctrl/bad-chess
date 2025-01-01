package main

import "errors"

type MatchId string

// TODO: Spectators  []*Client
type Match struct {
	LightPlayer *Client
	DarkPlayer  *Client

	Game *ChessGame
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
