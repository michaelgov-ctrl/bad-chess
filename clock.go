package main

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
	Done     chan<- struct{}
	lifeTime time.Duration
	started  time.Time
	elapsed  time.Duration
	state    ClockState
}

func NewClock(timeControl TimeControl) *Clock {
	clock := &Clock{
		Done:     make(chan<- struct{}),
		lifeTime: time.Duration(timeControl),
		started:  time.Now(),
		state:    running,
	}

	go clock.init()

	return clock
}

func (c *Clock) init() {
	ticker := time.NewTicker(50 * time.Millisecond)
	for range ticker.C {
		c.Lock()
		if c.lifeTime <= (c.elapsed + time.Since(c.started)) {
			c.state = expired
			close(c.Done)
			return
		}
		c.Unlock()
	}
}

func (c *Clock) Pause() {
	c.Lock()
	defer c.Unlock()

	if c.state != running {
		return
	}

	c.state = paused
	c.elapsed += time.Since(c.started)
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
