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

	c.currentMatch.TimeControl = joinEvent.TimeControl

	// this is probably slow
	m.matchesMu.Lock()
	defer m.matchesMu.Unlock()
	for matchId, match := range m.matches[joinEvent.TimeControl] {
		// no created match should ever be missing a light player since theyre added at creation
		if match.DarkPlayer == nil {
			m.matches[joinEvent.TimeControl][matchId].DarkPlayer = &Player{Client: c}
			c.currentMatch.ID = matchId
			c.currentMatch.Pieces = Dark
			err := m.matchMakingAddClientToMatch(c)
			if err != nil {
				return err
			}

			// both players should now be present to start game
			return match.Start(m.matchCleanupChan)
		}
	}

	c.currentMatch.ID = m.newMatch(joinEvent.TimeControl)
	c.currentMatch.Pieces = Light
	err := m.matchMakingAddClientToMatch(c)
	return err
}

func (m *Manager) MakeMoveHandler(event Event, c *Client) error {
	m.logger.Info("make move handler", "event", event, "client", *c)

	var moveEvent MakeMoveEvent
	if err := json.Unmarshal(event.Payload, &moveEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	m.matchesMu.RLock()
	defer m.matchesMu.RUnlock()
	if match, ok := m.matches[c.currentMatch.TimeControl][c.currentMatch.ID]; ok {
		clientPlayerColor := match.ClientPieceColor(c)
		if clientPlayerColor == NoColor || clientPlayerColor != c.currentMatch.Pieces {
			return fmt.Errorf("player pieces are borked")
		}

		if err := match.MakeMove(c.currentMatch.Pieces, moveEvent.Move); err != nil {
			return err
		}

		outgoingEvent, err := NewOutgoingEvent(EventPropagateMove, PropagateMoveEvent{
			PlayerColor: clientPlayerColor.String(),
			MoveEvent:   moveEvent,
		})
		if err != nil {
			return err
		}

		// send only to other player, should an accept be sent back to client?
		// egress is handled in (c *Client) writeMessages()
		match.MessagePlayers(outgoingEvent, oppositePlayer(clientPlayerColor))
		/*
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
		*/
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
			State:       Waiting,
		}

		go match.notifyIfStale(m.matchCleanupChan)

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
