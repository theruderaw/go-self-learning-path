# 11. Testing

Testing is a first-class part of the Go toolchain, requiring no third-party framework to be useful, though small libraries such as `testify` are common for more expressive assertions.

## `testing`

The standard library package underlying all Go tests:

```go
import "testing"
```

Test files end in `_test.go` and reside alongside the code they test, in the same package:

```
user.go
user_test.go
```

A test function must begin with `Test`, take `*testing.T`, and return nothing:

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

`t.Errorf` marks a failure but allows the test to continue; `t.Fatalf` marks a failure and stops immediately, used when continuing would only produce confusing follow-on failures — for instance, a nil pointer resulting from a failed setup step.

## Unit tests

A unit test verifies one function or method in isolation:

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

The idiomatic pattern for testing many input/output pairs without repeating the test body — a slice of cases is defined and looped over:

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

This is the most common shape a Go test takes.

## Subtests

`t.Run(name, func(t *testing.T) {...})` creates a named subtest — each appears individually in test output, can be selected specifically with `-run`, and a failing subtest does not stop sibling subtests from running:

```bash
go test -run TestAdd/negative_numbers ./...
```

## Test helpers

A function that sets up shared state for tests. Marking it with `t.Helper()` causes failure line numbers to point to the caller of the helper rather than the helper itself:

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

Handlers are tested directly by invoking them with a fake request/response, without a real server or network:

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

The package behind the preceding example — provides `httptest.NewRequest` (builds a fake request without a real socket), `httptest.NewRecorder` (captures a handler's response), and `httptest.NewServer` (starts a real local server for integration-style tests, such as testing an HTTP client):

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("mocked response"))
}))
defer server.Close()

resp, _ := http.Get(server.URL)
```

## `go test`

Runs the test suite:

```bash
go test ./...              # all packages, recursively
go test ./internal/user     # a single package
go test -v ./...            # verbose — shows each test name and result
go test -run TestAdd ./...  # only tests matching this name/pattern
go test -cover ./...         # shows coverage percentage
```

## Race detector

Detects data races (concurrent unsynchronized access) at runtime — important to run regularly on any code using goroutines:

```bash
go test -race ./...
```

It runs slower than a normal test pass due to added instrumentation, so it is often run in CI rather than on every local save, but it is typically part of the CI pipeline for any concurrent code.
