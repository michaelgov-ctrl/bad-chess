package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

func (m *Manager) MatchMakingHandler(event Event, c *Client) error {
	fmt.Printf("event: %v\nclient: %v\n", event, *c)

	// this is probably slow
	m.Lock()
	defer m.Unlock()
	for matchId, match := range m.matches {
		if match.LightPlayer == nil {
			m.matches[matchId].LightPlayer = c
			c.match = matchId
			return nil
		}

		if match.DarkPlayer == nil {
			m.matches[matchId].DarkPlayer = c
			c.match = matchId
			return nil
		}
	}

	matchId := m.newMatch()
	m.matches[matchId].LightPlayer = c
	c.match = matchId

	return nil
}

func (m *Manager) MakeMoveHandler(event Event, c *Client) error {
	fmt.Printf("event: %v\nclient: %v\n", event, *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	//validate move & move in game as client
	// *here*
	//validate move & move in game as client

	propEvent := PropagateMoveEvent{
		moveEvent,
	}

	data, err := json.Marshal(propEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal propagation event: %v", err)
	}

	outgoingEvent := Event{
		Payload: data,
		Type:    EventPropagateMove,
	}

	m.Lock()
	defer m.Unlock()
	if match, ok := m.matches[c.match]; ok {
		clientPlayerColor, err := match.ClientPlayerColor(c)
		if err != nil {
			c.egress <- Event{
				// TODO: payload should probably maybe be ErrorEvent
				Payload: []byte(`{error:"no color assigned for this match"}`),
				Type:    EventMatchError,
			}
			return err
		}

		// send only to other player, should an accept be sent back to client?
		// egress is handled in (c *Client) writeMessages()
		switch clientPlayerColor {
		case Light:
			if match.DarkPlayer != nil {
				match.DarkPlayer.egress <- outgoingEvent
			}
		case Dark:
			if match.LightPlayer != nil {
				match.LightPlayer.egress <- outgoingEvent
			}
		}
	} else {
		c.egress <- Event{
			// TODO: payload should probably maybe be ErrorEvent
			Payload: []byte(`{error:"client not assigned a match"}`),
			Type:    EventMatchError,
		}
		return errors.New("no match")
	}

	return nil
}

// TODO: this is probably fine for now, consider tossing an error on collision & retrying
func (m *Manager) newMatch() MatchId {
	matchId := MatchId(uuid.NewString())

	// right now newMatch() is only called in matchmaking which locks m
	if _, ok := m.matches[matchId]; ok {
		log.Println("uuid collision", matchId)
		matchId = m.newMatch()
	} else {
		match := &Match{
			ID:        matchId,
			ChessGame: newChessGame(),
		}

		go match.notifyWhenOver(m.matchCleanupChan)

		m.matches[matchId] = match
	}

	return matchId
}

/*
func (m *Manager) PropagateMoveHandler(event Event, c *Client) error {
	fmt.Printf("event: %v\nclient: %v\n", event, *c)
	if match, ok := m.matches[c.match]; ok {
		fmt.Printf("match: %v\n", *match)
	}
	return nil
}
*/
