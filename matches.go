package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/notnil/chess"
)

var (
	SupportedTimeControls = map[TimeControl]bool{
		TimeControl(5 * float64(time.Minute)):  true, // 5 minutes
		TimeControl(10 * float64(time.Minute)): true, // 10 minutes
		TimeControl(20 * float64(time.Minute)): true, // 20 minutes
	}
)

type MatchId string

type PieceColor int

const (
	Light PieceColor = iota
	Dark
	NoColor
)

type TimeControl time.Duration

type MatchList map[MatchId]*Match

type TimeControlMatchList map[TimeControl]MatchList

// TODO: Spectators  []*Client
type Match struct {
	ID MatchId

	TimeControl TimeControl

	LightPlayer *Client
	DarkPlayer  *Client

	Game *chess.Game
	Turn PieceColor
}

func (tc TimeControl) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(tc).String())
}

func (tc *TimeControl) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}

		if !SupportedTimeControls[TimeControl(tmp)] {
			return errors.New("unsupported time control")
		}

		*tc = TimeControl(tmp)
		return nil
	default:
		return errors.New("invalid time control")
	}
}

func (tc TimeControl) ToDuration() time.Duration {
	return time.Duration(tc)
}

func (m *Match) ClientPieceColor(client *Client) (PieceColor, error) {
	if m.LightPlayer != nil && m.LightPlayer == client {
		return Light, nil
	}

	if m.DarkPlayer != nil && m.DarkPlayer == client {
		return Dark, nil
	}

	return NoColor, errors.New("missing player")
}

func (m *Match) notifyWhenOver(ch chan<- ClientMatchInfo) {
	started := time.Now()
	waitTime := (m.TimeControl.ToDuration() * 2) + (15 * time.Second) // max wait time is for each players clock with a 15 second buffer
	for m.Game.Outcome() == chess.NoOutcome && time.Since(started) <= waitTime {
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("signaling to close", m.ID)
	ch <- ClientMatchInfo{
		ID:          m.ID,
		TimeControl: m.TimeControl,
	}
}
