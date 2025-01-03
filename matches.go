package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

const (
	LightWon = "1-0"
	DarkWon  = "0-1"
	Draw     = "1/2-1/2"
)

type TimeControl time.Duration

type MatchList map[MatchId]*Match

type TimeControlMatchList map[TimeControl]MatchList

type Player struct {
	Client *Client
	Clock  *Clock
}

// TODO: Spectators  []*Client
type Match struct {
	ID MatchId

	TimeControl TimeControl

	LightPlayer *Player
	DarkPlayer  *Player

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

func (m *Match) ClientPieceColor(client *Client) PieceColor {
	if m.LightPlayer != nil && m.LightPlayer.Client == client {
		return Light
	}

	if m.DarkPlayer != nil && m.DarkPlayer.Client == client {
		return Dark
	}

	return NoColor
}

type MatchOutcome struct {
	ID          MatchId
	TimeControl TimeControl
	Outcome     string
	Method      string
}

func (m *Match) notifyWhenOver(ch chan<- MatchOutcome) {
	ticker := time.NewTicker(500 * time.Millisecond)
	started, waitTime := time.Now(), (m.TimeControl.ToDuration()*2)+(15*time.Second) // max wait time is for each players clock with a 15 second buffer

	var outcome = MatchOutcome{
		ID:          m.ID,
		TimeControl: m.TimeControl,
	}
OUTER:
	for {
		select {
		case <-ticker.C:
			if m.Game.Outcome() != chess.NoOutcome || time.Since(started) >= waitTime {
				outcome.Outcome = m.Game.Outcome().String()
				outcome.Method = m.Game.Method().String()
				break OUTER
			}
		case <-safePlayerClockChannel(m.LightPlayer):
			outcome.Outcome = DarkWon
			outcome.Method = "flagged"
			break OUTER
		case <-safePlayerClockChannel(m.DarkPlayer):
			outcome.Outcome = LightWon
			outcome.Method = "flagged"
			break OUTER
		}
	}

	log.Println("signaling to close", m.ID)
	ch <- outcome
}

func (m *Match) MakeMove(pieces PieceColor, move string) error {
	if m.Turn != pieces {
		return errors.New("not players turn")
	}

	if err := m.Game.MoveStr(move); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	m.swapRunningClock(pieces)
	m.Turn = oppositePlayer(pieces)

	return nil
}

func (m *Match) swapRunningClock(pieces PieceColor) {
	switch pieces {
	case Light:
		m.LightPlayer.Clock.Pause()
		m.DarkPlayer.Clock.Start()
	case Dark:
		m.DarkPlayer.Clock.Pause()
		m.LightPlayer.Clock.Start()
	}
}

func (pc PieceColor) String() string {
	switch pc {
	case Light:
		return "light"
	case Dark:
		return "dark"
	default:
		return "no_color"
	}
}

func oppositePlayer(pieces PieceColor) PieceColor {
	switch pieces {
	case Light:
		return Dark
	case Dark:
		return Light
	default:
		return NoColor
	}
}

func safePlayerClockChannel(p *Player) <-chan time.Time {
	if p == nil || p.Clock == nil {
		return nil
	}

	return p.Clock.Done
}
