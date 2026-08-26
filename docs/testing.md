# 11. Testing

Testing is a first-class citizen in Go — no third-party framework needed to get real value, though a few small libraries (like `testify`) are common for nicer assertions.

## `testing`

The standard library package all Go tests are built on:

```go
import "testing"
```

Test files end in `_test.go` and live alongside the code they test, in the same package:

```
user.go
user_test.go
```

A test function must start with `Test`, take `*testing.T`, and have no return value:

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

`t.Errorf` marks the test failed but keeps running; `t.Fatalf` marks it failed and stops immediately (use this when continuing would just cause confusing follow-on failures, e.g. a nil pointer from a failed setup step).

## Unit tests

A unit test checks one function or method in isolation:

```go
func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

## Table-driven tests

The idiomatic Go pattern for testing many input/output pairs without repeating the same test body — a slice of cases, looped over:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", -2, 3, 1},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

This is by far the most common shape of a Go test — expect to write this pattern constantly.

## Subtests

`t.Run(name, func(t *testing.T) {...})` creates a named subtest — each shows up individually in test output, can be run selectively with `-run`, and failing subtests don't stop sibling subtests from executing:

```bash
go test -run TestAdd/negative_numbers ./...
```

## Test helpers

A function that sets up common state for tests. Mark it with `t.Helper()` so failure line numbers point to the *caller* of the helper, not the helper itself:

```go
func newTestUser(t *testing.T, name string) *User {
    t.Helper()
    u := &User{Name: name}
    if err := u.Validate(); err != nil {
        t.Fatalf("failed to create test user: %v", err)
    }
    return u
}

func TestSomething(t *testing.T) {
    u := newTestUser(t, "Alice")
    // ...
}
```

## HTTP handler tests

Test handlers directly by calling them with a fake request/response, no real server or network needed:

```go
func TestHelloHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/hello", nil)
    w := httptest.NewRecorder()

    helloHandler(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
    }

    body, _ := io.ReadAll(resp.Body)
    if string(body) != "hello\n" {
        t.Errorf("got body %q, want %q", body, "hello\n")
    }
}
```

## `httptest`

The package behind the example above — provides `httptest.NewRequest` (build a fake request without a real socket), `httptest.NewRecorder` (capture a handler's response), and `httptest.NewServer` (spin up a real local server for integration-style tests, e.g. when testing an HTTP client):

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("mocked response"))
}))
defer server.Close()

resp, _ := http.Get(server.URL)
```

## `go test`

Runs your tests:

```bash
go test ./...              # all packages, recursively
go test ./internal/user     # just one package
go test -v ./...            # verbose — show each test name and result
go test -run TestAdd ./...  # only tests matching this name/pattern
go test -cover ./...         # show coverage percentage
```

## Race detector

Catches data races (concurrent unsynchronized access) at runtime — critical to run regularly on any code using goroutines:

```bash
go test -race ./...
```

It's slower than a normal test run (extra instrumentation), so many teams run it in CI but not on every local save — but it should absolutely be part of your CI pipeline for any concurrent code.
