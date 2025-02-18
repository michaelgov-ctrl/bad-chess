package main

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventAssignedMatch     = "assigned_match"
	EventClockUpdate       = "clock_update"
	EventJoinMatchRequest  = "join_match"
	EventMakeMove          = "make_move"
	EventMatchOver         = "match_over"
	EventMatchStarted      = "match_started"
	EventMatchError        = "match_error"
	EventNewMatchRequest   = "new_match"
	EventPropagateMove     = "propagate_move"
	EventPropagatePosition = "propagate_position"
)

type JoinMatchEvent struct {
	TimeControl TimeControl `json:"time_control"`
}

type MakeMoveEvent struct {
	Move string `json:"move"`
	//Player string `json:"player"`
}

type PropagateMoveEvent struct {
	PlayerColor string `json:"player"`
	MoveEvent   MakeMoveEvent
}

type PropagatePositionEvent struct {
	PlayerColor string `json:"player"`
	FEN         string `json:"fen"`
}

type ErrorEvent struct {
	Error string `json:"error"`
}

type ClockUpdateEvent struct {
	ClockOwner    string `json:"clock_owner"`
	TimeRemaining string `json:"time_remaining"`
}

func NewOutgoingEvent(t string, evt any) (Event, error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return Event{}, fmt.Errorf("failed to marshal event: %v: %v", evt, err)
	}

	out := Event{
		Payload: data,
		Type:    t,
	}

	return out, nil
}

func NewErrorToEvent(errorType, msg string) (*Event, error) {
	data, err := json.Marshal(ErrorEvent{Error: msg})
	if err != nil {
		return nil, err
	}

	e := &Event{
		Type:    msg,
		Payload: data,
	}

	return e, nil
}
