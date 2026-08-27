# 6. HTTP / Web

This is where the preceding fundamentals converge into a working backend. Go's standard library is unusually complete here — a production API can be built without any external framework.

## `net/http`

The standard library package covering both HTTP servers and clients. For many backends, this package alone is sufficient — no Express/Flask equivalent is required.

```go
import "net/http"
```

## HTTP server

The minimal server:

```go
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, world!")
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Beyond a toy example, an explicit `*http.Server` is generally constructed so that timeouts can be configured:

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

## Handlers

A handler is anything satisfying the `http.Handler` interface (one method: `ServeHTTP`). `http.HandlerFunc` adapts a plain function so it can act as a handler:

```go
func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hi")
}

http.HandleFunc("/hello", hello) // wrapped as an http.HandlerFunc automatically
```

## `http.Request`

Carries all information about the incoming request:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    r.Method          // "GET", "POST", etc.
    r.URL.Path         // "/users/42"
    r.URL.Query()       // query string parameters
    r.Header.Get("Authorization")
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()
}
```

## `http.ResponseWriter`

Used to construct the response — headers first, then the status code, then the body, in that order, since headers cannot be modified after the body is written:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

## Routing

Go 1.22+'s built-in `http.ServeMux` supports method matching and path parameters natively, covering many routing needs without an external router:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("POST /users", createUser)

http.ListenAndServe(":8080", mux)
```

For more elaborate routing requirements, third-party routers (`chi`, `gorilla/mux`) remain common, though the standard library now covers most PWA-backend use cases independently.

## HTTP methods

```go
mux.HandleFunc("GET /items", listItems)
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("PUT /items/{id}", updateItem)
mux.HandleFunc("DELETE /items/{id}", deleteItem)
```

Handlers are matched to REST conventions: GET reads, POST creates, PUT/PATCH updates, DELETE removes.

## Status codes

```go
http.StatusOK                  // 200
http.StatusCreated              // 201
http.StatusNoContent            // 204
http.StatusBadRequest           // 400
http.StatusUnauthorized         // 401
http.StatusForbidden            // 403
http.StatusNotFound             // 404
http.StatusInternalServerError  // 500
```

```go
w.WriteHeader(http.StatusCreated)
```

## Headers

Reading and setting headers:

```go
r.Header.Get("Content-Type")
r.Header.Get("Authorization")

w.Header().Set("Content-Type", "application/json")
w.Header().Set("Cache-Control", "no-store")
```

## Query parameters

```go
// GET /search?q=go&limit=10
func handler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")           // "go"
    limit := r.URL.Query().Get("limit")    // "10" (a string, parsed separately)
}
```

## Path parameters

With Go 1.22+'s router:

```go
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintln(w, "user id:", id)
})
```

## Middleware

Middleware wraps a handler to add cross-cutting behavior — logging, authentication, panic recovery — without modifying the handler itself. The pattern is a function taking a handler and returning a new one.

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s took %v", r.Method, r.URL.Path, time.Since(start))
    })
}

handler := loggingMiddleware(mux)
http.ListenAndServe(":8080", handler)
```

Multiple middlewares are chained by nested wrapping: `logging(auth(recover(mux)))`.

## JSON APIs

Most PWA backends exchange JSON in both directions — covered fully in the next section, but the shape within a handler is:

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
```

## HTTP clients

For calling other APIs from the backend:

```go
resp, err := http.Get("https://api.example.com/data")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

body, _ := io.ReadAll(resp.Body)

// POST with a body:
resp, err = http.Post(url, "application/json", bytes.NewReader(jsonBytes))

// Full control:
req, _ := http.NewRequest("PUT", url, body)
req.Header.Set("Authorization", "Bearer "+token)
client := &http.Client{}
resp, err = client.Do(req)
```

## Timeouts

The zero-value `http.Client` has no timeout and can hang indefinitely on a stuck connection, so an explicit timeout is set:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

For per-request cancellation, a `context.Context` is paired with the request instead of relying on a blanket client timeout — covered in the Context section.
