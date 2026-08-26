# 7. JSON & Data

Nearly every request and response in a PWA backend passes through JSON. This section is about converting between Go structs and JSON reliably.

## `encoding/json`

The standard library package for all JSON work — no third-party library needed for typical use:

```go
import "encoding/json"
```

## Struct tags

Tags after a field tell `encoding/json` how to map that field to a JSON key:

```go
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email,omitempty"`   // omit if empty
    Password  string    `json:"-"`                  // never include in JSON
    CreatedAt time.Time `json:"created_at"`
}
```

Without a tag, the field is exported under its exact Go name (`Name`, not `name`) — which is why you almost always want explicit tags for a clean API. `omitempty` drops the field entirely from the output when it's the zero value. `-` excludes it unconditionally (useful for things like password hashes).

## Marshal

Converting a Go value **into** JSON:

```go
u := User{ID: 1, Name: "Alice", Email: "a@example.com"}

data, err := json.Marshal(u)
// data is []byte: {"id":1,"name":"Alice","email":"a@example.com","created_at":"0001-01-01T00:00:00Z"}

pretty, err := json.MarshalIndent(u, "", "  ") // indented for readability
```

In an HTTP handler, skip the intermediate `[]byte` and encode straight to the response:

```go
json.NewEncoder(w).Encode(u)
```

## Unmarshal

Converting JSON **into** a Go value, the field must be exported (capitalized) or `encoding/json` can't set it:

```go
data := []byte(`{"id":1,"name":"Alice"}`)

var u User
err := json.Unmarshal(data, &u) // note the &, Unmarshal needs a pointer
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

A full handler pattern to be reused constantly:

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

Use separate `Request` and `Response` structs rather than reusing database model directly — it keeps public API shape decoupled from your internal data model, so one can be changed without breaking the other.

## Validation

Go doesn't have built-in struct validation and ir is written explicitly (or a library like `go-playground/validator` once app grows). For most PWA backends, explicit manual validation is fine and easier to reason about:

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