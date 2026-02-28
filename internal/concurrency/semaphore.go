package concurrency

import (
	"context"
	"sync"
)

// Group manages concurrency groups with semaphores
type Group struct {
	mu         sync.Mutex
	semaphores map[string]*Semaphore
}

// NewGroup creates a new concurrency group manager
func NewGroup() *Group {
	return &Group{
		semaphores: make(map[string]*Semaphore),
	}
}

// Acquire acquires a slot in the named concurrency group
// maxParallel defines how many can run in parallel (0 = unlimited)
func (g *Group) Acquire(ctx context.Context, name string, maxParallel int) error {
	if maxParallel <= 0 {
		return nil // No limit
	}

	g.mu.Lock()
	sem, ok := g.semaphores[name]
	if !ok {
		sem = NewSemaphore(maxParallel)
		g.semaphores[name] = sem
	}
	g.mu.Unlock()

	return sem.Acquire(ctx)
}

// Release releases a slot in the named concurrency group
func (g *Group) Release(name string) {
	g.mu.Lock()
	sem, ok := g.semaphores[name]
	g.mu.Unlock()

	if ok {
		sem.Release()
	}
}

// Semaphore is a counting semaphore for limiting concurrency
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given capacity
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, n),
	}
}

// Acquire acquires a slot, blocking until one is available or context is cancelled
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a slot without blocking
// Returns true if acquired, false otherwise
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release releases a slot
func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
		// Already empty, ignore
	}
}

// Available returns the number of available slots
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// Capacity returns the total capacity
func (s *Semaphore) Capacity() int {
	return cap(s.ch)
}
