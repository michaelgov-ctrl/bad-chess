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
			m.matches[joinEvent.TimeControl][matchId].DarkPlayer = &Player{
				Client: c,
				Clock:  NewClock(joinEvent.TimeControl),
			}
			c.currentMatch = ClientMatchInfo{
				ID:          matchId,
				TimeControl: joinEvent.TimeControl,
				Pieces:      Dark,
			}
			return nil
		}
	}

	matchId := m.newMatch(joinEvent.TimeControl)
	m.matches[joinEvent.TimeControl][matchId].LightPlayer = &Player{
		Client: c,
		Clock:  NewClock(joinEvent.TimeControl),
	}
	c.currentMatch = ClientMatchInfo{
		ID:          matchId,
		TimeControl: joinEvent.TimeControl,
		Pieces:      Light,
	}

	/*
		m.matches[joinEvent.TimeControl][matchId].DarkPlayer = &Player{
			Clock: NewClock(joinEvent.TimeControl),
		}
	*/
	return nil
}

func (m *Manager) MakeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	m.Lock()
	defer m.Unlock()
	if match, ok := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]; ok {
		clientPlayerColor := match.ClientPieceColor(c)
		if clientPlayerColor == NoColor || clientPlayerColor != c.currentMatch.Pieces {
			return fmt.Errorf("player pieces are borked")
		}

		if err := match.MakeMove(c.currentMatch.Pieces, moveEvent.Move); err != nil {
			return err
		}

		propEvent := PropagateMoveEvent{
			PlayerColor: clientPlayerColor.String(),
			MoveEvent:   moveEvent,
		}

		data, err := json.Marshal(propEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal propagation event: %v", err)
		}

		outgoingEvent := Event{
			Payload: data,
			Type:    EventPropagateMove,
		}

		// send only to other player, should an accept be sent back to client?
		// egress is handled in (c *Client) writeMessages()
		switch clientPlayerColor {
		case Light:
			if match.DarkPlayer != nil {
				match.DarkPlayer.Client.egress <- outgoingEvent
			}
		case Dark:
			if match.LightPlayer != nil {
				match.LightPlayer.Client.egress <- outgoingEvent
			}
		}
	} else {
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
