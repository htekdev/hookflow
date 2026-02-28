package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(3)

	if sem.Capacity() != 3 {
		t.Errorf("Capacity() = %d, want 3", sem.Capacity())
	}

	if sem.Available() != 3 {
		t.Errorf("Available() = %d, want 3", sem.Available())
	}

	// Acquire all slots
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := sem.Acquire(ctx); err != nil {
			t.Errorf("Acquire() error = %v", err)
		}
	}

	if sem.Available() != 0 {
		t.Errorf("Available() = %d, want 0", sem.Available())
	}

	// Try acquire should fail
	if sem.TryAcquire() {
		t.Error("TryAcquire() should return false when full")
	}

	// Release one
	sem.Release()

	if sem.Available() != 1 {
		t.Errorf("Available() = %d, want 1", sem.Available())
	}

	// Try acquire should succeed now
	if !sem.TryAcquire() {
		t.Error("TryAcquire() should return true after release")
	}
}

func TestSemaphoreContextCancellation(t *testing.T) {
	sem := NewSemaphore(1)

	// Acquire the only slot
	ctx := context.Background()
	_ = sem.Acquire(ctx)

	// Try to acquire with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sem.Acquire(cancelCtx)
	if err != context.Canceled {
		t.Errorf("Acquire() error = %v, want context.Canceled", err)
	}
}

func TestSemaphoreTimeout(t *testing.T) {
	sem := NewSemaphore(1)

	// Acquire the only slot
	ctx := context.Background()
	_ = sem.Acquire(ctx)

	// Try to acquire with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(timeoutCtx)
	if err != context.DeadlineExceeded {
		t.Errorf("Acquire() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestSemaphoreConcurrent(t *testing.T) {
	sem := NewSemaphore(3)
	var maxConcurrent int32
	var currentConcurrent int32

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := context.Background()
			if err := sem.Acquire(ctx); err != nil {
				return
			}
			defer sem.Release()

			// Track concurrent executions
			current := atomic.AddInt32(&currentConcurrent, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&currentConcurrent, -1)
		}()
	}

	wg.Wait()

	if maxConcurrent > 3 {
		t.Errorf("Max concurrent = %d, want <= 3", maxConcurrent)
	}
}

func TestGroup(t *testing.T) {
	g := NewGroup()
	ctx := context.Background()

	// Acquire in group with limit 2
	if err := g.Acquire(ctx, "test-group", 2); err != nil {
		t.Errorf("Acquire() error = %v", err)
	}
	if err := g.Acquire(ctx, "test-group", 2); err != nil {
		t.Errorf("Acquire() error = %v", err)
	}

	// Third acquire should block, so we use a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := g.Acquire(timeoutCtx, "test-group", 2)
	if err != context.DeadlineExceeded {
		t.Errorf("Third Acquire() should timeout, got error = %v", err)
	}

	// Release one
	g.Release("test-group")

	// Now should be able to acquire
	if err := g.Acquire(ctx, "test-group", 2); err != nil {
		t.Errorf("Acquire() after release error = %v", err)
	}
}

func TestGroupUnlimited(t *testing.T) {
	g := NewGroup()
	ctx := context.Background()

	// maxParallel 0 means unlimited
	for i := 0; i < 100; i++ {
		if err := g.Acquire(ctx, "unlimited", 0); err != nil {
			t.Errorf("Acquire() error = %v", err)
		}
	}
}

func TestGroupMultipleGroups(t *testing.T) {
	g := NewGroup()
	ctx := context.Background()

	// Different groups should not interfere
	if err := g.Acquire(ctx, "group-a", 1); err != nil {
		t.Errorf("Acquire(group-a) error = %v", err)
	}
	if err := g.Acquire(ctx, "group-b", 1); err != nil {
		t.Errorf("Acquire(group-b) error = %v", err)
	}

	// group-a is full
	timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := g.Acquire(timeoutCtx, "group-a", 1)
	if err != context.DeadlineExceeded {
		t.Errorf("Acquire(group-a) should timeout, got error = %v", err)
	}

	// group-b is also full
	timeoutCtx2, cancel2 := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel2()

	err = g.Acquire(timeoutCtx2, "group-b", 1)
	if err != context.DeadlineExceeded {
		t.Errorf("Acquire(group-b) should timeout, got error = %v", err)
	}
}
