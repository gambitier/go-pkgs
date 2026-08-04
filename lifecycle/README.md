# lifecycle

Package `lifecycle` runs long-lived process components with **ordered start** and **reverse-order stop**.

Use it in a service `main` / bootstrap when you have several resources that must come up in dependency order and shut down cleanly on signal (or any context cancel)—for example a database client, telemetry, then an HTTP server.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/lifecycle@v0.1.0
```

## Component

A `Component` is any long-lived unit the process owns. You implement this interface in your service (or use a package that already does, such as `go-pkgs/mongodb`):

```go
type Component interface {
    // Name identifies the component in start/stop errors (e.g. "mongo", "http").
    Name() string

    // Start acquires resources or begins serving. Called once, in Add order.
    // The ctx is the same context passed to App.Run (usually cancelled on SIGINT/SIGTERM).
    Start(ctx context.Context) error

    // Stop releases resources. Called in reverse Add order on shutdown, and for
    // components that already started if a later Start fails.
    // The ctx is cancelled when stopTimeout (from New) expires.
    Stop(ctx context.Context) error
}
```

| Method | Responsibility | Examples |
|--------|----------------|----------|
| `Name` | Stable label for logs/errors | `"mongo"`, `"otel"`, `"http"` |
| `Start` | Connect, listen, init exporters | `mongo.Connect`, `fiber.Listen`, OTel `Init` |
| `Stop` | Drain and release | `Disconnect`, `ShutdownWithContext`, flush spans |

`Start` should return only after the component is ready for dependents (e.g. DB ping succeeded). For servers that block on accept, start listening in a goroutine and return once bind has succeeded (or return the listen error).

## App

```go
app := lifecycle.New(10 * time.Second) // per-Stop deadline; <=0 → 10s
app.Add(mongo, otel, http)             // start order = dependency order
err := app.Run(ctx)                    // Start all → wait ←ctx.Done() → Stop reverse
```

| API | Behavior |
|-----|----------|
| `New(stopTimeout)` | Builds an `App`. `stopTimeout` bounds **each** `Stop` so a hung shutdown cannot block exit forever. |
| `Add(...Component)` | Registers components. First added starts first and stops last. |
| `Run(ctx)` | Starts every component; on success blocks until `ctx` is done; then stops in reverse. On Start failure, stops already-started components and returns the joined error. |

Signal handling is **not** part of this package. In `cmd`, cancel `ctx` with `signal.NotifyContext`, then call `Run`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
return app.Run(ctx)
```

## Example

Implement a component, register it, run the app:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gambitier/go-pkgs/lifecycle"
)

// Worker is a Component that runs background work for the process lifetime.
type Worker struct {
	cancel context.CancelFunc
	done   chan struct{}
}

func (w *Worker) Name() string { return "worker" }

func (w *Worker) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				// do work
			}
		}
	}()
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	select {
	case <-w.done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("worker: stop: %w", ctx.Err())
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := lifecycle.New(10 * time.Second)
	app.Add(&Worker{})
	if err := app.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## Behavior summary

1. `Run` calls `Start` on each component in `Add` order.
2. If `Start` fails, already-started components get `Stop` (reverse order); `Run` returns.
3. If all start, `Run` blocks on `<-ctx.Done()`.
4. On cancel, each component gets `Stop` with a fresh context limited by `stopTimeout`, reverse order.
5. Stop errors are joined and returned; a single Stop failure does not skip the rest.

## Notes

- Stdlib only (`context`, `errors`, `fmt`, `time`).
- This package does not construct dependencies; your bootstrap creates components and passes them to `Add`.
- Pair with `go-pkgs/mongodb` for a ready-made Mongo `Component`, or implement `Component` for HTTP, workers, caches, etc.
