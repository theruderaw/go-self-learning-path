# 10. Database

For a PWA backend with real persistence, this is where user data, sessions, and app state actually live.

## `database/sql`

The standard library's generic SQL interface. It doesn't talk to any specific database on its own — it defines the API, and a separate **driver** package plugs in the actual database-specific logic:

```go
import (
    "database/sql"
    _ "github.com/lib/pq" // driver registers itself via side-effect import (the _)
)
```

The blank import (`_`) is intentional — you're not using any of the driver's exported names directly, just triggering its `init()` to register itself with `database/sql`.

## Connections

```go
db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// sql.Open doesn't actually connect yet — it just validates the arguments.
// Verify connectivity explicitly:
if err := db.Ping(); err != nil {
    log.Fatal(err)
}
```

`db` isn't a single connection — it's a connection *pool*, safe to share across goroutines and typically created once at startup and passed around your app.

## Queries

Three main methods, depending on what you expect back:

```go
db.Exec(...)     // for INSERT/UPDATE/DELETE — no rows returned
db.Query(...)    // for SELECT returning multiple rows
db.QueryRow(...) // for SELECT expected to return exactly one row
```

## `Exec`

For statements that don't return rows:

```go
result, err := db.Exec(
    "UPDATE users SET name = $1 WHERE id = $2",
    "Alice", 42,
)
if err != nil {
    log.Fatal(err)
}

rowsAffected, _ := result.RowsAffected()
```

(`$1`, `$2` is Postgres placeholder syntax; MySQL uses `?` instead — always use placeholders, never string-concatenate values into SQL, to avoid SQL injection.)

## `Query`

For SELECTs returning multiple rows — you must iterate and close:

```go
rows, err := db.Query("SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        log.Fatal(err)
    }
    users = append(users, u)
}
if err := rows.Err(); err != nil { // check for errors during iteration
    log.Fatal(err)
}
```

## `QueryRow`

For a SELECT expected to return exactly one row — no explicit loop or `Close()` needed:

```go
var u User
err := db.QueryRow("SELECT id, name FROM users WHERE id = $1", 42).Scan(&u.ID, &u.Name)
if err == sql.ErrNoRows {
    // no user found with that id
} else if err != nil {
    log.Fatal(err)
}
```

## `Scan`

Copies a row's columns into the addresses you provide, in the exact order the query selected them — the number and order of `Scan` arguments must match the SELECT columns exactly:

```go
rows.Scan(&u.ID, &u.Name, &u.Email)
```

Use `sql.NullString`, `sql.NullInt64`, etc. (or pointers) when a column might be `NULL`, since scanning `NULL` directly into a plain `string` returns an error.

## Transactions

Group multiple statements so they either all succeed or all roll back — critical whenever one logical operation touches more than one table/row:

```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

_, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, 1)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, 2)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

A common pattern is deferring a rollback that's a no-op if `Commit()` already succeeded:

```go
tx, _ := db.Begin()
defer tx.Rollback() // rolling back a committed tx is a harmless no-op

// ... tx.Exec calls ...

return tx.Commit()
```

## Prepared statements

Pre-compile a query once, then execute it repeatedly with different arguments — saves the database from re-parsing the SQL each time, useful for queries run in a loop or very frequently:

```go
stmt, err := db.Prepare("INSERT INTO logs (message) VALUES ($1)")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

for _, msg := range messages {
    stmt.Exec(msg)
}
```

Note that `db.Exec`/`db.Query` already use prepared statements under the hood per-call — `db.Prepare` is specifically for reusing the *same* compiled statement across many calls.

## Connection pooling

`database/sql` pools connections automatically, but the defaults are sometimes wrong for your workload — configure explicitly:

```go
db.SetMaxOpenConns(25)      // max simultaneous connections
db.SetMaxIdleConns(25)       // connections kept alive when idle
db.SetConnMaxLifetime(5 * time.Minute) // recycle connections periodically
```

Reasonable numbers depend on your database's own connection limits and your traffic — but leaving these unset entirely (unlimited open connections) is a common way to accidentally overwhelm a database under load.

## PostgreSQL driver

The two most common choices: `github.com/lib/pq` (older, simpler, pure-`database/sql`) and `github.com/jackc/pgx` (actively developed, faster, and can be used either via `database/sql` compatibility or its own richer native API). For a new project, `pgx` is generally the better default:

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/dbname")
```

Everything above (`Query`, `Exec`, `Scan`, transactions) works identically regardless of which driver you pick, since they all implement the same `database/sql` interface.
