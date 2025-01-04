package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/notnil/chess"
)

var (
	SupportedTimeControls = map[TimeControl]bool{
		TimeControl(1 * float64(time.Minute)):  true, // 1 minute
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

type MatchState int

const (
	Waiting MatchState = iota
	Started
	Over
)

// TODO: Spectators  []*Client
type Match struct {
	ID MatchId

	TimeControl TimeControl

	LightPlayer *Player
	DarkPlayer  *Player

	Game *chess.Game
	Turn PieceColor

	State MatchState
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

func (m *Match) Start(cleanupChan chan<- MatchOutcome) error {
	var outcome = MatchOutcome{ID: m.ID, TimeControl: m.TimeControl}
	if m.LightPlayer == nil || m.DarkPlayer == nil {
		outcome.Outcome = "0-0"
		outcome.Method = "abandonment"
		cleanupChan <- outcome
		return fmt.Errorf("failed to start match")
	}

	go m.notifyWhenOver(cleanupChan)

	m.State = Started
	m.LightPlayer.Clock = NewClock(m.TimeControl)
	m.DarkPlayer.Clock = NewClock(m.TimeControl)

	return nil
}

func (m *Match) notifyWhenOver(cleanupChan chan<- MatchOutcome) {
	var outcome = MatchOutcome{ID: m.ID, TimeControl: m.TimeControl}

	ticker := time.NewTicker(500 * time.Millisecond)
OUTER:
	for {
		select {
		case <-ticker.C:
			if m.Game.Outcome() != chess.NoOutcome {
				outcome.Outcome = m.Game.Outcome().String()
				outcome.Method = m.Game.Method().String()
				break OUTER
			}
			if m.State != Started {
				outcome.Outcome = "abandoned"
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

	m.State = Over
	cleanupChan <- outcome
}

func (m *Match) notifyIfStale(cleanupChan chan<- MatchOutcome) {
	ticker := time.NewTicker(500 * time.Millisecond)
	startTime, waitTime := time.Now(), (m.TimeControl.ToDuration()*2)+(30*time.Second) // max wait time is for each players clock with a 30 second buffer

	outcome := MatchOutcome{
		ID:          m.ID,
		TimeControl: m.TimeControl,
		Outcome:     "abandoned",
	}

	for range ticker.C {
		if time.Since(startTime) >= 20*time.Second && m.State == Waiting { // if the match hasnt started after 20 seconds kill it
			cleanupChan <- outcome
			return
		}

		if time.Since(startTime) >= waitTime && m.State != Over {
			cleanupChan <- outcome
			return
		}
	}
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

func (m *Match) MessagePlayers(event Event, players ...PieceColor) {
	for _, color := range players {
		switch color {
		case Light:
			if m.LightPlayer.Client != nil {
				m.LightPlayer.Client.egress <- event
			}
		case Dark:
			if m.DarkPlayer.Client != nil {
				m.DarkPlayer.Client.egress <- event
			}
		}
	}
}
