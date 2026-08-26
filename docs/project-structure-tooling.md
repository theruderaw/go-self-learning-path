# 12. Project Structure & Tooling

The last piece: how to lay out a real project, configure it, and use the tools that ship with Go itself.

## Go project layout

There's no framework-enforced structure like Rails or Django — but a widely-used convention for backend services looks roughly like this:

```
myapp/
├── go.mod
├── go.sum
├── cmd/
│   └── server/
│       └── main.go        # entry point, wires everything together
├── internal/
│   ├── user/
│   │   ├── handler.go       # HTTP handlers
│   │   ├── service.go       # business logic
│   │   ├── repository.go    # database access
│   │   └── user.go          # domain types
│   └── config/
│       └── config.go
├── migrations/
│   └── 0001_create_users.sql
└── README.md
```

`cmd/` holds entry points (you might have more than one binary — a server, a CLI migration tool, etc.). `internal/` holds your actual application code, organized by domain/feature rather than by technical layer (avoid a project-wide `controllers/`, `models/`, `services/` split — group by feature instead, it scales better).

## Internal packages

Any package under a directory literally named `internal/` can only be imported by code *inside* the same module — the Go toolchain enforces this at compile time, not just by convention:

```
myapp/
├── internal/
│   └── user/       // importable only from within myapp
└── pkg/
    └── validator/  // importable by other modules too, if you publish this one
```

Use `internal/` for anything that's an implementation detail of your app, and reserve a top-level `pkg/` (or no separate folder at all) only for code you genuinely intend other projects to import.

## Configuration

No built-in config framework — the idiomatic approach is a small `Config` struct populated from environment variables, often with a `.env` file for local development (loaded via a library like `github.com/joho/godotenv`, since Go doesn't read `.env` files natively):

```go
type Config struct {
    Port        string
    DatabaseURL string
    JWTSecret   string
}

func LoadConfig() (*Config, error) {
    cfg := &Config{
        Port:        getEnv("PORT", "8080"),
        DatabaseURL: os.Getenv("DATABASE_URL"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
    }
    if cfg.DatabaseURL == "" {
        return nil, errors.New("DATABASE_URL is required")
    }
    return cfg, nil
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

## Environment variables

```go
os.Getenv("PORT")               // "" if unset — no error
val, ok := os.LookupEnv("PORT")  // ok tells you whether it was actually set
os.Setenv("PORT", "8080")        // mostly useful in tests
```

Never hardcode secrets (database passwords, API keys, JWT secrets) — always load them from the environment, and keep `.env` out of version control via `.gitignore`.

## Logging

The standard library's `log` package is fine for simple needs:

```go
log.Println("server starting")
log.Printf("listening on port %s", port)
log.Fatal(err) // logs then calls os.Exit(1)
```

For a real backend, structured logging is worth the small extra setup — Go 1.21+ ships `log/slog` in the standard library, which gives you queryable key-value logs instead of plain strings:

```go
import "log/slog"

slog.Info("user created", "user_id", 42, "email", "a@example.com")
slog.Error("failed to save user", "error", err)
```

## `gofmt`

Auto-formats Go code to the one canonical style — there's no debate about tabs vs. spaces or brace placement in Go, because everyone just runs this:

```bash
gofmt -w .        # format all files in place
go fmt ./...       # equivalent, module-aware wrapper around gofmt
```

Most editors run this automatically on save. Commit only formatted code — CI often checks this and fails the build otherwise.

## `go vet`

Static analysis that catches suspicious code `gofmt` doesn't — things like a `Printf` call with mismatched format verbs, or a struct copied that shouldn't be:

```bash
go vet ./...
```

Cheap to run, catches real bugs, and is usually wired into CI right next to `go build` and `go test`.

## `go build`

Compiles your code into a binary, without running it:

```bash
go build ./cmd/server            # produces a "server" binary in the current dir
go build -o bin/myapp ./cmd/server # explicit output path/name
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server # cross-compile for Linux
```

## `go run`

Compiles and immediately runs — convenient for local development, but note it doesn't produce a reusable binary; use `go build` for that:

```bash
go run ./cmd/server
go run main.go
```

## `go install`

Compiles a binary and places it in `$GOPATH/bin` (or `$GOBIN`), typically used for installing CLI tools — either your own, or third-party ones — so they're available on your `$PATH`:

```bash
go install ./cmd/mytool
go install github.com/some/tool@latest
```

## `go test`

Covered in depth in the Testing section — included here because it's part of the same toolchain used constantly during day-to-day development, right alongside `build`, `run`, and `vet`:

```bash
go test ./...
```

A typical local/CI workflow strings several of these together:

```bash
go fmt ./...
go vet ./...
go test -race ./...
go build ./cmd/server
```
