package lifecycle_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gambitier/go-pkgs/lifecycle"
)

type stubComponent struct {
	name      string
	startErr  error
	stopErr   error
	started   *[]string
	stopped   *[]string
	mu        *sync.Mutex
	startHook func()
}

func (s *stubComponent) Name() string { return s.name }

func (s *stubComponent) Start(context.Context) error {
	if s.startHook != nil {
		s.startHook()
	}
	if s.startErr != nil {
		return s.startErr
	}
	s.mu.Lock()
	*s.started = append(*s.started, s.name)
	s.mu.Unlock()
	return nil
}

func (s *stubComponent) Stop(context.Context) error {
	s.mu.Lock()
	*s.stopped = append(*s.stopped, s.name)
	s.mu.Unlock()
	return s.stopErr
}

func TestAppStartsInOrderAndStopsInReverse(t *testing.T) {
	var (
		mu      sync.Mutex
		started []string
		stopped []string
	)

	app := lifecycle.New(time.Second)
	app.Add(
		&stubComponent{name: "a", started: &started, stopped: &stopped, mu: &mu},
		&stubComponent{name: "b", started: &started, stopped: &stopped, mu: &mu},
		&stubComponent{name: "c", started: &started, stopped: &stopped, mu: &mu},
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := join(started); got != "a,b,c" {
		t.Fatalf("start order = %q, want a,b,c", got)
	}
	if got := join(stopped); got != "c,b,a" {
		t.Fatalf("stop order = %q, want c,b,a", got)
	}
}

func TestAppStartFailureStopsAlreadyStarted(t *testing.T) {
	var (
		mu      sync.Mutex
		started []string
		stopped []string
	)

	app := lifecycle.New(time.Second)
	app.Add(
		&stubComponent{name: "a", started: &started, stopped: &stopped, mu: &mu},
		&stubComponent{name: "b", startErr: errors.New("boom"), started: &started, stopped: &stopped, mu: &mu},
		&stubComponent{name: "c", started: &started, stopped: &stopped, mu: &mu},
	)

	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected start error")
	}
	if !errors.Is(err, err) && err.Error() == "" {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := join(started); got != "a" {
		t.Fatalf("started = %q, want a", got)
	}
	if got := join(stopped); got != "a" {
		t.Fatalf("stopped = %q, want a", got)
	}
}

func join(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ","
		}
		out += p
	}
	return out
}
