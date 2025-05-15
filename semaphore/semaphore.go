package semaphore

import (
	"context"
)

type Semaphore struct {
	ch chan struct{}
}

func New(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.ch <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
		panic("semaphore released more than acquired")
	}
}
