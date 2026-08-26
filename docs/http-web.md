# 6. HTTP/Web 

This is crucial for a PWA backend. Go's standard library is complete to build a real production API without any framework.

## `net/http`

The standard library package for both HTTP servers and clients. For a lot of backends, this is all needed, no Express/Flask equivalent required.

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

For anything beyond a toy, a new `*http.Server` can be used so timeouts can be set:

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

A handler is anything satisfying the `http.Handler` interface (one method: `ServeHTTP`). `http.HandlerFunc` is an adapter that lets a plain function act as a handler:

```go
func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hi")
}

http.HandleFunc("/hello", hello) // wraps hello as an http.HandlerFunc automatically
```

## `http.Request`

Everything about the incoming request lives here:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    r.Method          // "GET", "POST", etc.
    r.URL.Path         // "/users/42"
    r.URL.Query()       // query string params
    r.Header.Get("Authorization")
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()
}
```

## `http.ResponseWriter`

How you write the response with headers first, then status code, then body, in that order (you can't change headers after writing the body):

```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

## Routing

Go 1.22+'s built-in `http.ServeMux` supports method matching and path parameters natively, so an external router is needed for maximum cases:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("POST /users", createUser)

http.ListenAndServe(":8080", mux)
```

For more complex routing needs, popular third-party routers (`chi`, `gorilla/mux`) are still common, but the standard library now covers most PWA-backend use cases on its own.

## HTTP methods

```go
mux.HandleFunc("GET /items", listItems)
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("PUT /items/{id}", updateItem)
mux.HandleFunc("DELETE /items/{id}", deleteItem)
```

Handlers matched to REST conventions: GET reads, POST creates, PUT/PATCH updates, DELETE removes.

## Status codes

```go
// 2xx Success
http.StatusOK                              // 200
http.StatusCreated                         // 201
http.StatusAccepted                        // 202
http.StatusNoContent                       // 204

// 3xx Redirection
http.StatusMovedPermanently                // 301
http.StatusFound                           // 302
http.StatusNotModified                     // 304

// 4xx Client Errors
http.StatusBadRequest                      // 400
http.StatusUnauthorized                    // 401
http.StatusForbidden                       // 403
http.StatusNotFound                        // 404
http.StatusMethodNotAllowed                // 405
http.StatusConflict                        // 409
http.StatusUnprocessableEntity             // 422
http.StatusTooManyRequests                 // 429

// 5xx Server Errors
http.StatusInternalServerError             // 500
http.StatusBadGateway                      // 502
http.StatusServiceUnavailable              // 503
http.StatusGatewayTimeout                  // 504
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
    limit := r.URL.Query().Get("limit")    // "10" (as a string — parsing needed)
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

A middleware wraps a handler to add cross-cutting behavior (logging, auth, recovery) without touching the handler itself. The pattern: a function that takes a handler and returns a new handler.

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

Chain multiple by wrapping repeatedly: `logging(auth(recover(mux)))`.

## JSON APIs

Most PWA backends are JSON in, JSON out — covered fully in the next section, but here's the shape in a handler:

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

For calling other APIs from backend:

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

Never use the zero-value `http.Client` in production as it has no timeout and will hang forever on a stuck connection:

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

For per-request cancellation (more flexible than a blanket client timeout), pair a request with a `context.Context` — covered in the Context section.