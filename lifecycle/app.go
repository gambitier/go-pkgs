package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Component is a long-lived process dependency started and stopped by App.
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// App starts components in order and stops them in reverse on shutdown or failure.
type App struct {
	components  []Component
	stopTimeout time.Duration
}

// New creates an App. stopTimeout bounds each Component.Stop during shutdown.
func New(stopTimeout time.Duration) *App {
	if stopTimeout <= 0 {
		stopTimeout = 10 * time.Second
	}
	return &App{stopTimeout: stopTimeout}
}

// Add registers components in start order (deps first).
func (a *App) Add(components ...Component) {
	a.components = append(a.components, components...)
}

// Run starts all components, waits for ctx cancellation, then stops everything.
func (a *App) Run(ctx context.Context) error {
	started := make([]Component, 0, len(a.components))

	for _, c := range a.components {
		if err := c.Start(ctx); err != nil {
			stopErr := a.stopAll(started)
			return errors.Join(fmt.Errorf("start %s: %w", c.Name(), err), stopErr)
		}
		started = append(started, c)
	}

	<-ctx.Done()
	return a.stopAll(started)
}

func (a *App) stopAll(components []Component) error {
	var joined error
	for i := len(components) - 1; i >= 0; i-- {
		c := components[i]
		stopCtx, cancel := context.WithTimeout(context.Background(), a.stopTimeout)
		err := c.Stop(stopCtx)
		cancel()
		if err != nil {
			joined = errors.Join(joined, fmt.Errorf("stop %s: %w", c.Name(), err))
		}
	}
	return joined
}
