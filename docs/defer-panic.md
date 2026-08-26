# 4. Defer and Panic

Go doesn't have try/catch/finally. `defer`, `panic`, and `recover` cover the same ground, but with a different philosophy: errors are values to be checked, and panics are reserved for truly exceptional, unrecoverable situations.

## `defer`

Schedules a function call to run right before the surrounding function returns — no matter how it returns (normal return, early return, or even a panic):

```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // guaranteed to run when readFile exits

    // ... use f ...
    return nil
}
```

This is the idiomatic way to guarantee cleanup — closing files, unlocking mutexes, closing database connections is doneright next to the code that acquires the resource, so it's impossible to forget.

Arguments to a deferred call are evaluated immediately, but the call itself runs later:

```go
x := 1
defer fmt.Println("x was", x) // captures x's value NOW (1), even though it prints later
x = 2
```

## Defer execution order

Multiple `defer` calls in the same function run in **LIFO** order — last deferred, first executed:

```go
func demo() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// prints: 3, 2, 1
```

Works in the reverse order of acquisition. Emulates a stack.

## `panic`

`panic` immediately stops normal execution of the current function, runs any deferred calls, then does the same for its caller, and so on up the stack — until either something `recover`s, or the program crashes with a stack trace:

```go
func mustPositive(n int) int {
    if n < 0 {
        panic("n must be positive")
    }
    return n
}
```

Panics are for programmer errors and truly unrecoverable states — not for expected failure conditions like "file not found" or "invalid user input." Those should be regular `error` returns.

## `recover`

`recover` stops a panic in progress and lets the program continue. It only works when called directly inside a deferred function:

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

A common real-world use: an HTTP server wrapping every handler in middleware that recovers from panics, so one bad request doesn't crash the whole server:

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

The rule of thumb: **`error` for anything expected**, **`panic` for anything that should never happen**.

| Situation | Use |
| --- | --- |
| File doesn't exist | `error` |
| Invalid user input | `error` |
| Database connection failed | `error` |
| Index out of a slice's bounds due to a real bug | `panic` (this happens automatically) |
| Nil pointer dereference | `panic` (automatic) |
| Programmer passed an impossible config that should've been caught earlier | `panic`, often via a package-level `init()` check |