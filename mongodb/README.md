# mongodb

Thin MongoDB connect + lifecycle component glue over the official driver.

## What / why

Standardize URI connect, primary ping, and `lifecycle.Component` Start/Stop across services. This is **not** a data-access layer — repositories still use `*mongo.Database` and `go.mongodb.org/mongo-driver/v2` directly.

Indexes and schema changes belong in service-owned **golang-migrate** migrations, not here.

## Install

```bash
export GOPRIVATE=github.com/gambitier/*
go get github.com/gambitier/go-pkgs/mongodb@v0.1.0
```

## Usage

```go
import (
  "github.com/gambitier/go-pkgs/lifecycle"
  "github.com/gambitier/go-pkgs/mongodb"
)

mongoComp := mongodb.NewComponent(uri, database, serverSelectionTimeout)
app := lifecycle.New(10 * time.Second)
app.Add(mongoComp, /* otel, http */)
// after Start: mongoComp.DB()
```

## Important notes

- Depends on mongo-driver v2 only (implements `lifecycle.Component` by structural typing; no import of `lifecycle`).
- Independent of `logging` / `observability`.
- Optional `serverSelectionTimeout` (`<= 0` leaves driver default).
