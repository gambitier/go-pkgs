# mongodb

Package `mongodb` is **thin process glue** over the official MongoDB Go driver v2: connect with a primary ping, and a ready-made `lifecycle.Component` for Start/Stop.

This is **not** a data-access layer. Repositories keep using `*mongo.Database`, collections, and BSON from `go.mongodb.org/mongo-driver/v2` directly. Indexes and schema changes belong in **service-owned migrations** (e.g. golang-migrate JSON), not here.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/mongodb@v0.1.0
```

## When to use

- Standardize URI connect + primary ping across services
- Register Mongo as a `lifecycle` component (`Name` / `Start` / `Stop`) without copying connect code
- Avoid inventing a second Mongo API on top of the driver

## Connect

```go
client, db, err := mongodb.Connect(ctx, uri, "identity", 5*time.Second)
if err != nil {
    return err
}
defer client.Disconnect(ctx)

_ = db.Collection("users")
```

| Parameter | Behavior |
|-----------|----------|
| `uri` | Passed to `options.Client().ApplyURI` |
| `database` | `client.Database(database)` |
| `serverSelectionTimeout` | Applied when `> 0`; otherwise driver default |
| ping | Primary read preference, 10s timeout; on failure disconnects and returns error |

## Component (lifecycle)

`Component` implements the `lifecycle.Component` interface by structural typing (this module does **not** import `go-pkgs/lifecycle`).

```go
import (
    "github.com/gambitier/go-pkgs/lifecycle"
    "github.com/gambitier/go-pkgs/mongodb"
)

mongoComp := mongodb.NewComponent(uri, database, serverSelectionTimeout)

app := lifecycle.New(10 * time.Second)
app.Add(mongoComp /*, otel, http */)
// Run starts mongo first when listed first

// After Start succeeds (e.g. inside a later component):
db := mongoComp.DB()
```

| Method | Behavior |
|--------|----------|
| `Name()` | `"mongo"` |
| `Start(ctx)` | Calls `Connect`; stores client/DB |
| `Stop(ctx)` | `Disconnect`; clears state |
| `DB()` / `Client()` | Return handles after Start; **panic** if not started |

## What this package does not do

- Repository base types or query helpers
- Index creation (`CreateIndexes`) — use migrations
- Health probes beyond connect-time ping
- Logging or OpenTelemetry instrumentation of commands

## Notes

- Depends on `go.mongodb.org/mongo-driver/v2` only among go-pkgs siblings.
- Independent of `logging` / `observability`.
- Tags: `mongodb/vX.Y.Z`.
- Pair with `go-pkgs/lifecycle` in bootstrap; keep domain repos in the service.
