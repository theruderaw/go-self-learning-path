# 5. Concurrency

Main benefit of Go. For a PWA backend, this is used for things like handling many requests at once, running background jobs, and fanning work out to multiple workers.

## Goroutines

A goroutine is a lightweight, Go-managed thread. It is started just by prefixing a function call with `go`:

```go
func sayHello() {
    fmt.Println("hello")
}

go sayHello()          // runs concurrently, doesn't block
go func() {              // anonymous goroutine
    fmt.Println("world")
}()
```

They're cheap  and thousands can be run without issue — but `main()` doesn't wait for goroutines to finish on its own. If `main()` returns, all goroutines are killed mid-flight, so the progrAM NEEDS to wait for them (see `sync.WaitGroup` below).

## Channels

A channel is a typed pipe for goroutines to send values to each other safely, without manual locking:

```go
ch := make(chan int)

go func() {
    ch <- 42 // send
}()

value := <-ch // receive (blocks until something is sent)
```

Channels are the concurrency-safe way to move data *between* goroutines, rather than sharing memory and hoping you locked it correctly.

## Buffered / unbuffered channels

An **unbuffered** channel (`make(chan int)`) blocks the sender until a receiver is ready — it's a synchronous handoff. A **buffered** channel (`make(chan int, 3)`) lets you send up to its capacity without blocking:

```go
unbuffered := make(chan int)
buffered := make(chan int, 3)

buffered <- 1 // doesn't block, buffer has room
buffered <- 2
buffered <- 3
// buffered <- 4 would block — buffer is full
```

Use unbuffered channels when using a strict handshake; use buffered producer and consumer speed need to be decoupled.

## Channel direction

A channel can restrict its parameter to send-only or receive-only, which the compiler enforces and is useful for documenting intent in function signatures:

```go
func send(ch chan<- int, v int) { // send-only
    ch <- v
}

func receive(ch <-chan int) int { // receive-only
    return <-ch
}
```

## Closing channels

`close(ch)` signals "no more values will be sent." Receivers can detect this:

```go
close(ch)
v, ok := <-ch // ok is false once the channel is closed and drained
```

Rule: only the **sender** should close a channel, never the receiver, and a channel closed twice panics.

## `range` over channels

Reading from a channel in a loop until it's closed:

```go
ch := make(chan int)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // must close, or range blocks forever
}()

for v := range ch {
    fmt.Println(v)
}
```

## `select`

Waits on multiple channel operations at once, proceeding with whichever is ready first. This is Go's version of a `switch` for channels:

```go
select {
case v := <-ch1:
    fmt.Println("from ch1:", v)
case v := <-ch2:
    fmt.Println("from ch2:", v)
case <-time.After(2 * time.Second):
    fmt.Println("timed out")
default:
    fmt.Println("nothing ready right now") // makes select non-blocking
}
```

This is how timeouts and cancellations around channels work.

## `sync.WaitGroup`

Waits for a group of goroutines to finish and is the standard way to make `main()` (or any function) wait:

```go
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        fmt.Println("worker", n)
    }(i) // pass i explicitly to avoid the classic loop-variable bug
}

wg.Wait() // blocks until all 3 call Done()
```

`Add(1)` before starting each goroutine, `Done()` (usually deferred) inside it, `Wait()` to block until the count reaches zero.

## `sync.Mutex`

Protects shared state from concurrent access when you genuinely need to share memory rather than pass it over a channel:

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

Without the lock, two goroutines incrementing `count` at the same time can lose updates — that's a race condition (next section).

## Race conditions

A race condition happens when two goroutines access the same memory concurrently, and at least one is writing, without synchronization. The result is undefined, sometimes it works, sometimes data gets silently corrupted.

```go
counter := 0
var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter++ // NOT safe — read-modify-write isn't atomic
    }()
}
wg.Wait()
fmt.Println(counter) // often NOT 1000
```

It is fixed with a mutex, a channel, or `sync/atomic` for simple counters. Go ships a built-in tool to catch these via race detector in the Testing section.

## Basic concurrency patterns

**Worker pool** — a fixed number of goroutines pulling work off a shared channel:

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

**Fan-out/fan-in** — multiple goroutines processing in parallel, results collected on one channel. **Pipeline** — chaining stages together, each stage's output channel feeding the next stage's input. These are just combinations of the primitives above — once you're comfortable with goroutines, channels, `select`, and `WaitGroup`, these patterns fall out naturally.