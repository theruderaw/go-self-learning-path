# 13. pgx (PostgreSQL Driver, In Depth)

Section 10 covered `database/sql` in general, using `pgx` only through its `database/sql`-compatible mode. This section covers `pgx`'s own **native API** — a separate, richer way of communicating with Postgres that most new Go backends use directly, since it is faster and exposes Postgres-specific features unavailable through `database/sql`.

## Two ways to use pgx

```go
// Mode 1: database/sql compatibility layer
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)
db, _ := sql.Open("pgx", connString) // functions exactly like section 10

// Mode 2: pgx's own native API — no database/sql involved
import "github.com/jackc/pgx/v5/pgxpool"
pool, _ := pgxpool.New(context.Background(), connString)
```

Mode 2 is the subject of this manual. It is not a drop-in replacement for `database/sql` — method names, return types, and pooling all differ — but it is generally the better default for a new project, offering greater speed and direct access to Postgres-specific types and features.

## Installing

```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

## Connection strings

pgx accepts a standard Postgres connection URL:

```go
connString := "postgres://user:password@localhost:5432/mydb?sslmode=disable"
```

## `pgxpool` — the connection pool

Unlike `database/sql`, where `sql.Open` provides a pool implicitly, native pgx requires the pool to be created explicitly with `pgxpool`. This is what application code interacts with directly — a single raw connection is rarely used:

```go
import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    ctx := context.Background()

    pool, err := pgxpool.New(ctx, connString)
    if err != nil {
        log.Fatalf("unable to create connection pool: %v", err)
    }
    defer pool.Close()

    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("unable to reach database: %v", err)
    }
}
```

Every pgx method takes a `context.Context` as its first argument — there is no context-less variant, unlike `database/sql`'s separate `db.Query`/`db.QueryContext` split. This reflects an assumption that cancellation and timeout support are always wanted.

## Pool configuration

The pool is configured via a `pgxpool.Config`, parsed from the connection string and then adjusted:

```go
config, err := pgxpool.ParseConfig(connString)
if err != nil {
    log.Fatal(err)
}

config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
```

## Queries — `Query`, `QueryRow`, `Exec`

The same three-way split as `database/sql`, with a `ctx` required on every call:

```go
// Exec — no rows returned
_, err := pool.Exec(ctx,
    "UPDATE users SET name = $1 WHERE id = $2",
    "Alice", 42,
)

// QueryRow — exactly one row expected
var name string
err = pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 42).Scan(&name)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        // no user found
    }
}

// Query — multiple rows
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        log.Fatal(err)
    }
    fmt.Println(id, name)
}
if err := rows.Err(); err != nil {
    log.Fatal(err)
}
```

Note `pgx.ErrNoRows` rather than `sql.ErrNoRows` — a frequent source of small bugs when porting code between the two modes.

## `pgx.CollectRows` — scanning without a manual loop

A significant improvement over `database/sql`: rather than a hand-written `for rows.Next() { ... }` loop, rows can be scanned directly into a slice of structs:

```go
type User struct {
    ID   int
    Name string
}

rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}

users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
if err != nil {
    log.Fatal(err)
}
// users is now []User, fully populated
```

`pgx.RowToStructByName` matches columns to struct fields by name (with `db:"column_name"` struct tags added if names differ), removing most of the boilerplate present in section 10's manual `Scan` loop. `pgx.RowToStructByPos` matches by column order instead of name, and `pgx.CollectOneRow` handles the single-row case.

## Transactions

The same concept as `database/sql`, again with `ctx` required throughout:

```go
tx, err := pool.Begin(ctx)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx) // a no-op if already committed

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, 1)
if err != nil {
    return err
}

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, 2)
if err != nil {
    return err
}

return tx.Commit(ctx)
```

pgx also provides `pool.BeginFunc`/`pool.BeginTxFunc`, a convenience wrapper that commits automatically on success and rolls back automatically if the supplied function returns an error or panics:

```go
err := pool.BeginFunc(ctx, func(tx pgx.Tx) error {
    if _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, 1); err != nil {
        return err
    }
    _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, 2)
    return err
})
```

## Batch — sending multiple queries in one round trip

`pgx.Batch` queues several statements and sends them to Postgres together, reducing network round trips — useful when many independent statements must run efficiently:

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "first")
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "second")
batch.Queue("INSERT INTO logs (message) VALUES ($1)", "third")

results := pool.SendBatch(ctx, batch)
defer results.Close()

for i := 0; i < 3; i++ {
    if _, err := results.Exec(); err != nil {
        log.Fatal(err)
    }
}
```

`database/sql` has no equivalent mechanism — this feature is specific to pgx.

## `COPY` — fast bulk inserts

For loading many rows at once, pgx exposes Postgres's `COPY` protocol, substantially faster than issuing individual `INSERT` statements:

```go
rows := [][]any{
    {1, "Alice"},
    {2, "Bob"},
    {3, "Carol"},
}

count, err := pool.CopyFrom(
    ctx,
    pgx.Identifier{"users"},
    []string{"id", "name"},
    pgx.CopyFromRows(rows),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("inserted", count, "rows")
```

## Postgres-specific types (`pgtype`)

Because pgx communicates with Postgres directly, rather than through the generic `database/sql` abstraction, it natively supports Postgres types with no clean equivalent in plain Go — arrays, JSON/JSONB, ranges, UUIDs, and `NULL`-aware wrappers:

```go
import "github.com/jackc/pgx/v5/pgtype"

var tags []string
err := pool.QueryRow(ctx, "SELECT tags FROM posts WHERE id = $1", 1).Scan(&tags)
// Postgres text[] maps directly to a Go []string

var settings map[string]any
err = pool.QueryRow(ctx, "SELECT settings FROM users WHERE id = $1", 1).Scan(&settings)
// jsonb maps directly to a Go map

var nickname pgtype.Text // Postgres-aware "nullable string"
err = pool.QueryRow(ctx, "SELECT nickname FROM users WHERE id = $1", 1).Scan(&nickname)
if nickname.Valid {
    fmt.Println(nickname.String)
}
```

This is among the strongest reasons to prefer native pgx over the `database/sql` compatibility mode — array and JSON columns become directly accessible with well-typed values, instead of requiring manual `[]byte` handling.

## Error handling — `pgconn.PgError`

pgx surfaces detailed Postgres error information (SQLSTATE codes, constraint names) rather than a generic error string, allowing a unique-constraint violation to be distinguished from any other failure:

```go
import "github.com/jackc/pgx/v5/pgconn"

_, err := pool.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", "taken@example.com")
if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            http.Error(w, "email already exists", http.StatusConflict)
            return
        }
    }
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
}
```

## Where the pool is used in an application

The same pattern as `database/sql`: the pool is created once at startup, then passed (or wrapped in a repository struct) into whatever needs it — a new pool is never created per request.

```go
type UserRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    var u User
    err := r.pool.QueryRow(ctx, "SELECT id, name FROM users WHERE id = $1", id).
        Scan(&u.ID, &u.Name)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrUserNotFound
    }
    return &u, err
}
```

```go
// main.go
pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

userRepo := NewUserRepository(pool)
```

## `database/sql` mode vs native pgx — comparison

| | `database/sql` + pgx driver | native pgx (`pgxpool`) |
|---|---|---|
| API style | generic, works with any SQL database | Postgres-specific |
| Performance | good | faster — no abstraction overhead |
| Postgres types (arrays, JSONB, UUID) | manual handling | native support via `pgtype` |
| Batch queries, `COPY` | not available | built in |
| Swappable to another database later | yes | no — locked to Postgres |
| Ecosystem tools expecting `*sql.DB` | works directly | requires the `stdlib` compatibility shim |

For a PWA backend committed to Postgres — the common case — native pgx via `pgxpool` is generally the better starting choice. The `database/sql` compatibility mode is reserved for situations requiring integration with a tool that expects a standard `*sql.DB`, such as certain migration tools, ORMs, or observability libraries.
