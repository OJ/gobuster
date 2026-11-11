package libgobuster

import (
	"context"
	"sync"
)

// PauseController handles pause/resume for workers
type PauseController struct {
	mu       sync.RWMutex
	paused   bool
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

// NewPauseController creates a new unpaused controller
func NewPauseController() *PauseController {
	return &PauseController{
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
	}
}

// Pause pauses the controller
func (p *PauseController) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		p.paused = true
		close(p.pauseCh)
	}
}

// Resume unpauses the controller
func (p *PauseController) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.paused {
		p.paused = false
		p.pauseCh = make(chan struct{})
		close(p.resumeCh)
		p.resumeCh = make(chan struct{})
	}
}

// Toggle flips pause state and returns the new state
func (p *PauseController) Toggle() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.paused {
		p.paused = false
		p.pauseCh = make(chan struct{})
		close(p.resumeCh)
		p.resumeCh = make(chan struct{})
	} else {
		p.paused = true
		close(p.pauseCh)
	}
	return p.paused
}

// IsPaused returns true if paused
func (p *PauseController) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.paused
}

// Wait blocks until resumed or context is cancelled
func (p *PauseController) Wait(ctx context.Context) error {
	p.mu.RLock()
	if !p.paused {
		p.mu.RUnlock()
		return nil
	}
	resumeCh := p.resumeCh
	p.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-resumeCh:
		return nil
	}
}
