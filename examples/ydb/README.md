# YDB Example

This example demonstrates how to use ent with [YDB](https://ydb.tech/) database.

## Prerequisites

- Running YDB instance (local or remote)
- Go 1.22+

## Running YDB Locally

You can run YDB locally using Docker:

```bash
docker run -d --rm --name ydb-local \
  -p 2135:2135 -p 2136:2136 -p 8765:8765 \
  -e GRPC_TLS_PORT=2135 \
  -e GRPC_PORT=2136 \
  -e MON_PORT=8765 \
  -e YDB_USE_IN_MEMORY_PDISKS=true \
  cr.yandex/yc/yandex-docker-local-ydb:latest
```

## Running the Example

```bash
# Set the connection string (adjust for your YDB instance)
export YDB_CONNECTION_STRING="grpc://localhost:2136/local"

# Run the example
go run example.go
```

## Key Features Demonstrated

1. **YDB Driver** - Using the ent YDB driver with ydb-go-sdk
2. **Automatic Retries** - Using `WithRetryOptions()` for YDB-specific retry handling
3. **Schema Creation** - Creating tables in YDB using ent migrations
4. **CRUD Operations** - Create, Read, Update, Delete operations with YDB

## Schema

This example uses a simple TV series database with three entities:

- **Series** - TV series (title, info, release date)
- **Season** - Seasons of a series (title, first/last aired dates)
- **Episode** - Episodes in a season (title, air date, duration)

## Retry Handling

YDB requires special retry handling for transient errors. The ent YDB driver 
automatically handles retries using ydb-go-sdk's retry package. You can 
customize retry behavior using `WithRetryOptions()`:

```go
// Create with custom retry options
_, err := client.Series.Create().
    SetTitle("Breaking Bad").
    SetInfo("A high school chemistry teacher turned meth producer").
    SetReleaseDate(time.Date(2008, 1, 20, 0, 0, 0, 0, time.UTC)).
    WithRetryOptions(retry.WithIdempotent(true)).
    Save(ctx)

// Query with retry options
series, err := client.Series.Query().
    Where(series.TitleContains("Bad")).
    WithRetryOptions(retry.WithIdempotent(true)).
    All(ctx)
```
