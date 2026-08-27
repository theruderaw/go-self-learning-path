# 7. JSON & Data

Nearly every request and response in a PWA backend passes through JSON. This section covers converting between Go structs and JSON.

## `encoding/json`

The standard library package handling JSON encoding and decoding — no third-party library is required for typical use:

```go
import "encoding/json"
```

## Struct tags

A tag following a field tells `encoding/json` how to map that field to a JSON key:

```go
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email,omitempty"`   // omitted when empty
    Password  string    `json:"-"`                  // never included in JSON
    CreatedAt time.Time `json:"created_at"`
}
```

Without a tag, a field is exported under its exact Go name (`Name`, not `name`), which is why explicit tags are used for a clean external API. `omitempty` drops a field entirely from the output when it holds the zero value. `-` excludes a field unconditionally — useful for values such as password hashes.

## Marshal

Converts a Go value **into** JSON:

```go
u := User{ID: 1, Name: "Alice", Email: "a@example.com"}

data, err := json.Marshal(u)
// data is []byte: {"id":1,"name":"Alice","email":"a@example.com","created_at":"0001-01-01T00:00:00Z"}

pretty, err := json.MarshalIndent(u, "", "  ") // indented for readability
```

In an HTTP handler, the intermediate `[]byte` is usually skipped, encoding directly to the response:

```go
json.NewEncoder(w).Encode(u)
```

## Unmarshal

Converts JSON **into** a Go value — the target fields must be exported (capitalized) or `encoding/json` cannot set them:

```go
data := []byte(`{"id":1,"name":"Alice"}`)

var u User
err := json.Unmarshal(data, &u) // note the &; Unmarshal requires a pointer
```

From an HTTP request body, a decoder is used instead of reading the whole body first:

```go
var u User
if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
}
```

## JSON request/response handling

A recurring handler pattern:

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // ... validate, save to DB ...

    resp := UserResponse{ID: 1, Name: req.Name, Email: req.Email}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

Separate `Request` and `Response` structs are used rather than reusing the database model directly, keeping the public API shape decoupled from the internal data model — one can change without breaking the other.

## Validation

Go has no built-in struct validation; it is written explicitly, or handled with a library such as `go-playground/validator` as an application grows. Manual validation is generally sufficient for most PWA backends:

```go
func (r CreateUserRequest) Validate() error {
    if strings.TrimSpace(r.Name) == "" {
        return errors.New("name is required")
    }
    if !strings.Contains(r.Email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := req.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // ... proceed ...
}
```

If `go-playground/validator` is adopted later, the same idea applies through struct tags (`validate:"required,email"`) rather than hand-written checks.
