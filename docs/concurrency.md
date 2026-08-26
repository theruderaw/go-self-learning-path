# 5. Concurrency

This is the area Go is most known for. In a PWA backend, it governs how many requests are handled at once, how background jobs run, and how work is distributed across workers.

## Goroutines

A goroutine is a lightweight, Go-managed thread, started by prefixing a function call with `go`:

```go
func sayHello() {
    fmt.Println("hello")
}

go sayHello()          // runs concurrently, does not block
go func() {              // anonymous goroutine
    fmt.Println("world")
}()
```

Goroutines are cheap enough that thousands can run simultaneously without issue. `main()` does not wait for goroutines to finish on its own — if `main()` returns, all goroutines are terminated mid-execution, which is why a synchronization mechanism such as `sync.WaitGroup` is needed to wait for completion.

## Channels

A channel is a typed pipe used by goroutines to exchange values safely, without manual locking:

```go
ch := make(chan int)

go func() {
    ch <- 42 // send
}()

value := <-ch // receive (blocks until something is sent)
```

Channels provide a concurrency-safe way to move data between goroutines, as an alternative to sharing memory directly and relying on manual locking.

## Buffered / unbuffered channels

An **unbuffered** channel (`make(chan int)`) blocks the sender until a receiver is ready — a synchronous handoff. A **buffered** channel (`make(chan int, 3)`) allows sends up to its capacity without blocking:

```go
unbuffered := make(chan int)
buffered := make(chan int, 3)

buffered <- 1 // does not block, buffer has room
buffered <- 2
buffered <- 3
// buffered <- 4 would block — buffer is full
```

Unbuffered channels enforce a strict handshake; buffered channels decouple producer and consumer speed to some degree.

## Channel direction

A channel parameter can be restricted to send-only or receive-only, a restriction enforced by the compiler and useful for documenting intent:

```go
func send(ch chan<- int, v int) { // send-only
    ch <- v
}

func receive(ch <-chan int) int { // receive-only
    return <-ch
}
```

## Closing channels

`close(ch)` signals that no further values will be sent. Receivers can detect this:

```go
close(ch)
v, ok := <-ch // ok becomes false once the channel is closed and drained
```

By convention, only the sender closes a channel; a receiver never does, and a channel is never closed more than once, which would cause a panic.

## `range` over channels

Reads from a channel repeatedly until it is closed:

```go
ch := make(chan int)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // required, or range blocks indefinitely
}()

for v := range ch {
    fmt.Println(v)
}
```

## `select`

Waits on multiple channel operations simultaneously, proceeding with whichever becomes ready first — analogous to a `switch` for channels:

```go
select {
case v := <-ch1:
    fmt.Println("from ch1:", v)
case v := <-ch2:
    fmt.Println("from ch2:", v)
case <-time.After(2 * time.Second):
    fmt.Println("timed out")
default:
    fmt.Println("nothing ready right now") // makes the select non-blocking
}
```

This mechanism implements timeouts and cancellation around channel-based work.

## `sync.WaitGroup`

Waits for a group of goroutines to complete — the standard way to make `main()` or any function block until concurrent work is finished:

```go
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        fmt.Println("worker", n)
    }(i) // i is passed explicitly to avoid the classic loop-variable bug
}

wg.Wait() // blocks until all three goroutines call Done()
```

`Add(1)` is called before starting each goroutine, `Done()` (usually deferred) is called inside it, and `Wait()` blocks until the internal counter reaches zero.

## `sync.Mutex`

Protects shared state from concurrent access, for situations where memory is genuinely shared rather than passed over a channel:

```go
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

Without the lock, two goroutines incrementing `count` simultaneously can lose updates — a race condition, covered next.

## Race conditions

A race condition occurs when two goroutines access the same memory concurrently and at least one is writing, without synchronization. The outcome is undefined: sometimes correct, sometimes silently corrupted.

```go
counter := 0
var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter++ // not safe — read-modify-write is not atomic
    }()
}
wg.Wait()
fmt.Println(counter) // often not 1000
```

The fix is a mutex, a channel, or `sync/atomic` for simple counters. Go includes a built-in detector for this class of bug, described in the Testing section.

## Basic concurrency patterns

**Worker pool** — a fixed number of goroutines pulling work from a shared channel:

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        results <- j * 2
    }
}

jobs := make(chan int, 100)
results := make(chan int, 100)

for w := 1; w <= 3; w++ {
    go worker(w, jobs, results)
}

for j := 1; j <= 5; j++ {
    jobs <- j
}
close(jobs)
```

**Fan-out/fan-in** distributes work across multiple goroutines and collects results on a single channel. **Pipeline** chains stages together, with each stage's output channel feeding the next stage's input. Both patterns are combinations of the primitives above — goroutines, channels, `select`, and `WaitGroup`.