# 8. Context

`context` propagates cancellation, deadlines, and request-scoped values through a call chain — essential in a web backend, where a client may disconnect mid-request or a downstream call may need a timeout.

## `context.Context`

An interface representing the context of an operation — carrying a cancellation signal, an optional deadline, and optional key-value data. It is passed as the first argument to any function that should respect cancellation:

```go
func doWork(ctx context.Context) error {
    // checks ctx.Done() or passes ctx along to anything accepting one
}
```

Every `http.Request` already carries one, accessible via `r.Context()`.

## `context.Background`

The root context — not derived from anything, never cancelled, with no deadline. Every context chain originates here (or with `context.TODO()`, a placeholder used when the correct context is not yet determined):

```go
ctx := context.Background()
```

This is used mainly at the top of a program — in `main`, or in tests. Inside HTTP handlers, `r.Context()` is used instead, since it is cancelled automatically if the client disconnects.

## `context.WithCancel`

Creates a child context that can be cancelled manually:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // called regardless of whether cancellation happens early, to release resources

go func() {
    time.Sleep(2 * time.Second)
    cancel() // triggers ctx.Done()
}()

<-ctx.Done()
fmt.Println("cancelled:", ctx.Err())
```

## `context.WithTimeout`

Behaves like `WithCancel` but cancels automatically after a duration — commonly used for outbound calls such as database queries or requests to other services:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := http.DefaultClient.Do(req)
// a request exceeding 5 seconds is cancelled, and err is set
```

`context.WithDeadline` serves the same purpose using an absolute `time.Time` instead of a duration.

## Request-scoped context

Request-scoped values — such as a user ID extracted by authentication middleware — can be attached to a context and read further down the call chain:

```go
type ctxKey string
const userIDKey ctxKey = "userID"

// in auth middleware:
ctx := context.WithValue(r.Context(), userIDKey, 42)
r = r.WithContext(ctx)
next.ServeHTTP(w, r)

// in a handler further down:
userID := r.Context().Value(userIDKey).(int)
```

A private, unexported key type is used (rather than a plain string) to prevent collisions with context values set by other packages. `context.Value` is reserved for request-scoped metadata such as authentication info or trace IDs, not for passing ordinary function arguments.

## Cancellation / deadlines

Long-running work checks `ctx.Done()` periodically, or passes the context to anything that natively understands it:

```go
func longRunningTask(ctx context.Context) error {
    for i := 0; i < 1000000; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err() // context.Canceled or context.DeadlineExceeded
        default:
            // performs a chunk of work
        }
    }
    return nil
}
```

Database drivers (`database/sql`) and `net/http` clients accept a context directly and handle this internally — for example, `db.QueryContext(ctx, ...)` or `http.NewRequestWithContext(ctx, ...)` — so the context is typically threaded through rather than checked manually. Any function performing I/O — network, database, file, or slow computation — accepts a `context.Context` as its first parameter, giving the caller control over how long it is willing to wait.
