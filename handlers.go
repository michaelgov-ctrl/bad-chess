package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/notnil/chess"
)

func (m *Manager) MatchMakingHandler(event Event, c *Client) error {
	m.logger.Info("match making handler", "event", event, "client", *c)

	var joinEvent JoinMatchEvent
	if err := json.Unmarshal(event.Payload, &joinEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// this is probably slow
	m.Lock()
	defer m.Unlock()
	for matchId, match := range m.matches[joinEvent.TimeControl] {
		// no created match should ever be missing a light player since theyre added at creation
		if match.DarkPlayer == nil {
			m.matches[joinEvent.TimeControl][matchId].DarkPlayer = c
			c.currentMatch = ClientMatchInfo{
				ID:          matchId,
				TimeControl: joinEvent.TimeControl,
				Pieces:      Dark,
			}
			return nil
		}
	}

	matchId := m.newMatch(joinEvent.TimeControl)
	m.matches[joinEvent.TimeControl][matchId].LightPlayer = c
	c.currentMatch = ClientMatchInfo{
		ID:          matchId,
		TimeControl: joinEvent.TimeControl,
		Pieces:      Light,
	}
	return nil
}

func (m *Manager) MakeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

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
	if match, ok := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]; ok {
		clientPlayerColor, err := match.ClientPieceColor(c)
		if err != nil || clientPlayerColor != c.currentMatch.Pieces {
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
func (m *Manager) newMatch(timeControl TimeControl) MatchId {
	matchId := MatchId(uuid.NewString())

	// right now newMatch() is only called in matchmaking which locks m
	if _, ok := m.matches[timeControl][matchId]; ok {
		m.logger.Error("uuid collision", "MatchId", matchId)
		matchId = m.newMatch(timeControl)
	} else {
		match := &Match{
			ID:          matchId,
			TimeControl: timeControl,
			Game:        chess.NewGame(),
			Turn:        Light,
		}

		go match.notifyWhenOver(m.matchCleanupChan)

		m.matches[timeControl][matchId] = match
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
