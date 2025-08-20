// Package run implements an actor-runner with deterministic teardown. It is
// somewhat similar to package errgroup, except it does not require actor
// goroutines to understand context semantics. This makes it suitable for use in
// more circumstances; for example, goroutines which are handling connections
// from net.Listeners, or scanning input from a closable io.Reader.
package runner

import (
	"context"
	"log"
	"time"
)

// Group collects actors (functions) and runs them concurrently.
// When one actor (function) returns, all actors are interrupted.
// The zero value of a Group is useful.
type Group struct {
	// actors is the list of actors (functions) to run.
	actors []actor
	// ShutdownTimeout is the maximum amount of time to wait for all actors to
	// stop. If zero, the default is 5 seconds.
	ShutdownTimeout time.Duration
}

// Add an actor (function) to the group. Each actor must be pre-emptable by an
// interrupt function. That is, if interrupt is invoked, execute should return.
// Also, it must be safe to call interrupt even after execute has returned.
//
// The first actor (function) to return interrupts all running actors.
// The error is passed to the interrupt functions, and is returned by Run.
func (g *Group) Add(execute func() error, interrupt func(error)) {
	g.actors = append(g.actors, actor{execute, interrupt})
}

// Run all actors (functions) concurrently.
// When the first actor returns, all others are interrupted.
// Run only returns when all actors have exited.
// Run returns the error returned by the first exiting actor.
func (g *Group) Run() error {
	if len(g.actors) == 0 {
		return nil
	}

	// Run each actor.
	errors := make(chan error, len(g.actors))
	for _, a := range g.actors {
		go func(a actor) {
			errors <- a.execute()
		}(a)
	}

	// Wait for the first actor to stop.
	err := <-errors

	// Signal all actors to stop.
	for _, a := range g.actors {
		a.interrupt(err)
	}

	if g.ShutdownTimeout == 0 {
		g.ShutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.ShutdownTimeout)
	defer cancel()

	// Wait for all actors to stop with timeout control
	for i := 1; i < cap(errors); i++ {
		select {
		case <-errors:
		case <-ctx.Done():
			// Timeout, return the original error.
			log.Printf("wait remaining actors exit timeout: %v", ctx.Err())
			return err
		}
	}

	// Return the original error.
	return err
}

// actor is a function that can be interrupted.
type actor struct {
	execute   func() error
	interrupt func(error)
}
