# 8. Context

`context` is how Go propagates cancellation, deadlines, and request-scoped values through a call chain — essential for a web backend where a client can disconnect mid-request, or a downstream call needs a timeout.

## `context.Context`

An interface representing "the context of this operation" — it carries a cancellation signal, an optional deadline, and optional key-value data, and gets passed as the *first* argument to any function that should respect cancellation:

```go
func doWork(ctx context.Context) error {
    // check ctx.Done() or pass ctx along to anything that accepts one
}
```

Every `http.Request` already carries one: `r.Context()`.

## `context.Background`

The root context — not derived from anything, never cancelled, no deadline. Every context chain starts here (or with `context.TODO()`, used as a placeholder when you're not yet sure what context to use):

```go
ctx := context.Background()
```

You'll mostly use this at the very top of your program (in `main`, or in tests); inside HTTP handlers you should use `r.Context()` instead, since it's automatically cancelled if the client disconnects.

## `context.WithCancel`

Creates a child context you can cancel manually:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // always call cancel to release resources, even if you didn't cancel early

go func() {
    time.Sleep(2 * time.Second)
    cancel() // triggers ctx.Done()
}()

<-ctx.Done()
fmt.Println("cancelled:", ctx.Err())
```

## `context.WithTimeout`

Like `WithCancel`, but auto-cancels after a duration — the most common one you'll use for outbound calls (database queries, HTTP requests to other services):

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := http.DefaultClient.Do(req)
// if the request takes longer than 5s, it's cancelled and err is set
```

There's also `context.WithDeadline`, which is the same idea but with an absolute `time.Time` instead of a duration.

## Request-scoped context

You can attach request-scoped values (like a user ID pulled from auth middleware) to a context, and read them further down the call chain:

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

Use a private, unexported key type (not a plain string) to avoid collisions with context values set by other packages. `context.Value` should be reserved for request-scoped metadata like auth info or trace IDs — not for passing regular function arguments; that's what actual parameters are for.

## Cancellation / deadlines

The pattern for respecting cancellation inside long-running work — check `ctx.Done()` periodically, or pass the context to anything that natively understands it:

```go
func longRunningTask(ctx context.Context) error {
    for i := 0; i < 1000000; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err() // context.Canceled or context.DeadlineExceeded
        default:
            // do a chunk of work
        }
    }
    return nil
}
```

Database drivers (`database/sql`) and `net/http` clients accept a context directly and handle this for you — e.g. `db.QueryContext(ctx, ...)` or `http.NewRequestWithContext(ctx, ...)` — so most of the time you're just threading the context through rather than writing the `select` yourself. The core habit worth building: any function that does I/O (network, database, file, or a slow computation) should accept a `context.Context` as its first parameter, so callers control how long they're willing to wait.