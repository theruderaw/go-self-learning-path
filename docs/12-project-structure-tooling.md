# 12. Project Structure & Tooling

The final piece: how a real project is laid out, configured, and built using the tools shipped with Go itself.

## Go project layout

There is no framework-enforced structure comparable to Rails or Django, but a widely used convention for backend services looks roughly like this:

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

`cmd/` holds entry points — a project may have more than one binary, such as a server and a separate CLI migration tool. `internal/` holds the application's actual code, organized by domain or feature rather than by technical layer; a project-wide split into `controllers/`, `models/`, `services/` is generally avoided in favor of grouping by feature, which scales better.

## Internal packages

Any package located under a directory literally named `internal/` can only be imported by code within the same module — enforced by the Go toolchain at compile time, not merely by convention:

```
myapp/
├── internal/
│   └── user/       // importable only from within myapp
└── pkg/
    └── validator/  // importable by other modules too, if this one is published
```

`internal/` is used for anything that is an implementation detail of the application, while a top-level `pkg/` (or no separate folder at all) is reserved for code genuinely intended for import by other projects.

## Configuration

No built-in configuration framework exists — the idiomatic approach is a small `Config` struct populated from environment variables, often with a `.env` file for local development, loaded via a library such as `github.com/joho/godotenv` since Go does not read `.env` files natively:

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
os.Getenv("PORT")               // "" if unset — produces no error
val, ok := os.LookupEnv("PORT")  // ok indicates whether the variable was actually set
os.Setenv("PORT", "8080")        // primarily useful in tests
```

Secrets — database passwords, API keys, JWT secrets — are never hardcoded; they are loaded from the environment, and `.env` is excluded from version control via `.gitignore`.

## Logging

The standard library's `log` package suffices for simple needs:

```go
log.Println("server starting")
log.Printf("listening on port %s", port)
log.Fatal(err) // logs, then calls os.Exit(1)
```

For a production backend, structured logging is generally worth the additional setup — Go 1.21+ includes `log/slog` in the standard library, producing queryable key-value logs rather than plain strings:

```go
import "log/slog"

slog.Info("user created", "user_id", 42, "email", "a@example.com")
slog.Error("failed to save user", "error", err)
```

## `gofmt`

Automatically formats Go code to a single canonical style, eliminating debate over tabs versus spaces or brace placement, since the same formatter is run universally:

```bash
gofmt -w .        # formats all files in place
go fmt ./...       # equivalent, module-aware wrapper around gofmt
```

Most editors run this automatically on save. Only formatted code is committed — CI frequently enforces this and fails the build otherwise.

## `go vet`

Static analysis catching suspicious code that `gofmt` does not — such as a `Printf` call with mismatched format verbs, or a struct copied in a way that shouldn't be:

```bash
go vet ./...
```

It is inexpensive to run and catches real bugs, typically wired into CI alongside `go build` and `go test`.

## `go build`

Compiles code into a binary without running it:

```bash
go build ./cmd/server            # produces a "server" binary in the current directory
go build -o bin/myapp ./cmd/server # explicit output path/name
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server # cross-compilation for Linux
```

## `go run`

Compiles and immediately runs the program — convenient during local development, though it does not produce a reusable binary; `go build` is used for that:

```bash
go run ./cmd/server
go run main.go
```

## `go install`

Compiles a binary and places it in `$GOPATH/bin` (or `$GOBIN`), typically used to install CLI tools — either project-specific or third-party — so they become available on `$PATH`:

```bash
go install ./cmd/mytool
go install github.com/some/tool@latest
```

## `go test`

Covered in depth in the Testing section, and included here because it is part of the same toolchain used throughout day-to-day development, alongside `build`, `run`, and `vet`:

```bash
go test ./...
```

A typical local or CI workflow chains several of these commands together:

```bash
go fmt ./...
go vet ./...
go test -race ./...
go build ./cmd/server
```
