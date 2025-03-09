package game

import (
	"sync"
	"time"
)

type ClockState int

const (
	running ClockState = iota
	paused
	expired
)

type Clock struct {
	sync.Mutex
	Done     <-chan time.Time
	lifeTime time.Duration
	started  time.Time
	elapsed  time.Duration
	state    ClockState
}

func NewClock(timeControl TimeControl) *Clock {
	doneChan := make(chan time.Time)

	clock := &Clock{
		Done:     doneChan,
		lifeTime: time.Duration(timeControl),
		started:  time.Now(),
		state:    running,
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range ticker.C {
			clock.Lock()
			if clock.state == running {
				clock.elapsed += time.Since(clock.started)
				clock.started = time.Now()
			}

			if clock.lifeTime <= clock.elapsed {
				clock.state = expired
				doneChan <- time.Now()
				close(doneChan)
				return
			}
			clock.Unlock()
		}
	}()

	return clock
}

func (c *Clock) Pause() {
	c.Lock()
	defer c.Unlock()

	if c.state != running {
		return
	}

	c.elapsed += time.Since(c.started)
	c.state = paused
}

func (c *Clock) Start() {
	c.Lock()
	defer c.Unlock()

	if c.state != paused {
		return
	}

	c.state = running
	c.started = time.Now()
}

func (c *Clock) TimeRemaining() time.Duration {
	return c.lifeTime - c.elapsed
}
