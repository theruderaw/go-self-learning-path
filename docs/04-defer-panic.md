# 4. Defer & Panic

Go has no try/catch/finally. `defer`, `panic`, and `recover` cover similar ground under a different philosophy: errors are values checked explicitly, while panics are reserved for genuinely unrecoverable situations.

## `defer`

Schedules a function call to run immediately before the surrounding function returns, regardless of how it returns — normal return, early return, or panic:

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // runs when readFile exits, no matter the exit path

    // ... use f ...
    return nil
}
```

This is the standard mechanism for guaranteeing cleanup — closing files, unlocking mutexes, closing database connections — placed directly next to the code that acquires the resource.

Arguments to a deferred call are evaluated immediately, though the call itself executes later:

```go
x := 1
defer fmt.Println("x was", x) // captures x's value now (1), even though printing happens later
x = 2
```

## Defer execution order

Multiple `defer` calls within one function execute in **LIFO** order — last deferred, first executed:

```go
func demo() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// prints: 3, 2, 1
```

This behaves like a stack, which is why resources acquired in sequence are released in reverse order — typically the correct order for cleanup.

## `panic`

`panic` halts normal execution of the current function immediately, runs any deferred calls, then propagates the same behavior up through its caller, and so on — until something calls `recover`, or the program terminates with a stack trace:

```go
func mustPositive(n int) int {
    if n < 0 {
        panic("n must be positive")
    }
    return n
}
```

Panics are reserved for programmer errors and truly unrecoverable states, not for expected failure conditions such as a missing file or invalid input — those are represented with a regular `error` return.

## `recover`

`recover` stops an in-progress panic, allowing execution to continue. It only has an effect when called directly inside a deferred function:

```go
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    result = a / b // panics if b == 0
    return
}
```

A frequent real-world application: HTTP middleware wrapping every handler with a recovery step, so a single bad request does not crash the entire server:

```go
func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v", err)
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

## `panic` vs `error`

The governing rule: `error` for anything expected, `panic` for anything that should never occur.

| Situation | Mechanism |
|---|---|
| File does not exist | `error` |
| Invalid user input | `error` |
| Database connection failed | `error` |
| Index out of a slice's bounds due to a genuine bug | `panic` (occurs automatically) |
| Nil pointer dereference | `panic` (automatic) |
| Configuration that should have been caught earlier and makes further execution impossible | `panic`, often triggered via a package-level `init()` check |

Reaching for `panic` to signal that a request was invalid is generally a sign that an `error` return is the correct tool instead.
