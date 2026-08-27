# 10. Database

For a PWA backend with real persistence, this is where user data, sessions, and application state reside.

## `database/sql`

The standard library's generic SQL interface. It does not communicate with any specific database on its own — it defines the API, while a separate **driver** package supplies the database-specific implementation:

```go
import (
    "database/sql"
    _ "github.com/lib/pq" // driver registers itself via a side-effect import (the _)
)
```

The blank import (`_`) is deliberate — none of the driver's exported names are used directly; only its `init()` function, which registers it with `database/sql`, is triggered.

## Connections

```go
db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// sql.Open does not connect immediately — it only validates its arguments.
// Connectivity is verified explicitly:
if err := db.Ping(); err != nil {
    log.Fatal(err)
}
```

`db` represents a connection *pool*, not a single connection. It is safe to share across goroutines and is typically created once at startup and passed throughout the application.

## Queries

Three methods, depending on the expected result:

```go
db.Exec(...)     // for INSERT/UPDATE/DELETE — no rows returned
db.Query(...)    // for SELECT returning multiple rows
db.QueryRow(...) // for SELECT expected to return exactly one row
```

## `Exec`

For statements that do not return rows:

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

(`$1`, `$2` is Postgres placeholder syntax; MySQL uses `?`. Placeholders are always used instead of string-concatenating values into SQL, to prevent SQL injection.)

## `Query`

For SELECTs returning multiple rows — iteration and closing are both required:

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
if err := rows.Err(); err != nil { // checks for errors that occurred during iteration
    log.Fatal(err)
}
```

## `QueryRow`

For a SELECT expected to return exactly one row — no explicit loop or `Close()` is needed:

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

Copies a row's columns into the provided addresses, in the exact order the query selected them — the number and order of `Scan` arguments must match the SELECT columns precisely:

```go
rows.Scan(&u.ID, &u.Name, &u.Email)
```

`sql.NullString`, `sql.NullInt64`, or pointers are used when a column may be `NULL`, since scanning `NULL` directly into a plain `string` produces an error.

## Transactions

Group multiple statements so they either all succeed or all roll back — necessary whenever one logical operation touches more than one table or row:

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

A common pattern defers a rollback that becomes a no-op once `Commit()` has already succeeded:

```go
tx, _ := db.Begin()
defer tx.Rollback() // rolling back a committed tx is a harmless no-op

// ... tx.Exec calls ...

return tx.Commit()
```

## Prepared statements

Pre-compiles a query once, then executes it repeatedly with different arguments, avoiding re-parsing the SQL on each execution — useful for queries run in a loop or very frequently:

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

`db.Exec`/`db.Query` already use prepared statements internally per call; `db.Prepare` is specifically for reusing the same compiled statement across many calls.

## Connection pooling

`database/sql` pools connections automatically, though the defaults are sometimes unsuited to a given workload:

```go
db.SetMaxOpenConns(25)      // maximum simultaneous connections
db.SetMaxIdleConns(25)       // connections retained while idle
db.SetConnMaxLifetime(5 * time.Minute) // periodic connection recycling
```

Appropriate values depend on the database's own connection limits and the application's traffic, but leaving these unset (unlimited open connections) is a common cause of a database becoming overwhelmed under load.

## PostgreSQL driver

The two most common choices are `github.com/lib/pq` (older, simpler, purely a `database/sql` driver) and `github.com/jackc/pgx` (actively developed, faster, usable either through `database/sql` compatibility or its own richer native API). For a new project, `pgx` is generally the better default:

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/dbname")
```

Everything described above (`Query`, `Exec`, `Scan`, transactions) works identically regardless of which driver is chosen, since both implement the same `database/sql` interface.
