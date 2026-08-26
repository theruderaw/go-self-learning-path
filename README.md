# Go Self-Learning Documentation

The goal of this learning document is to give developers a practical foundation in Go. Each stage introduces concepts that are immediately reinforced through a project.

## Learning Path

1. [Go Basics](docs/go-basics.md)
2. [Structs, Methods & Interfaces](docs/structs-methods-interfaces.md)
3. [Packages & Modules](docs/packages-modules.md)
4. [Defer & Panic](docs/defer-panic.md)
5. [Concurrency](docs/concurrency.md)
6. [HTTP / Web](docs/http-web.md)
7. [JSON & Data](docs/json-data.md)
8. [Context](docs/context.md)

## Learning Projects

### 1. CLI Task Manager

Build a command-line task manager with basic CRUD and persistence.

**Concepts**

- Go modules
- Variables and constants
- Functions
- Structs
- Methods
- Value vs pointer receivers
- Pointers
- Slices
- Maps
- Interfaces
- Error handling
- Packages
- File I/O
- JSON
- Basic testing

**Checkpoint:** Persistent CLI application.

---

### 2. File Organizer CLI

Build a CLI that organizes files in a directory based on their extensions.

**Concepts**

- `os`
- `filepath`
- Directory traversal
- File metadata
- Command-line arguments
- Maps
- Error handling
- Packages
- Practical filesystem operations

**Checkpoint:** A useful filesystem utility.

---

### 3. Log Analyzer CLI

Build a program that reads server/application logs and produces statistics.

**Concepts**

- File reading
- `bufio`
- String parsing
- Maps
- Structs
- Sorting
- Custom functions
- Error handling
- Basic benchmarking

**Checkpoint:** A data-processing CLI.

---

### 4. HTTP Server

Build a small HTTP server using Go's standard library.

**Concepts**

- `net/http`
- HTTP methods
- Request/response handling
- Routing
- JSON
- HTTP status codes
- Headers
- Middleware
- Query parameters
- Path parameters
- Context
- Basic concurrency

**Checkpoint:** First Go web server.

---

### 5. URL Shortener API

Build a small API that accepts URLs and generates short IDs.

Example:

```
POST /shorten
GET  /abc123
GET  /stats/abc123
```

**Concepts**

- REST API design
- JSON request/response bodies
- HTTP handlers
- HTTP middleware
- Routing
- In-memory state
- Concurrent access
- `sync.Mutex` / `sync.RWMutex`
- HTTP errors
- API testing

**Checkpoint:** Small but real API with concurrent state.

---

### 6. CLI Task Manager → API Task Manager

Take the original Task Manager and turn it into an HTTP API.

**Concepts**

- API architecture
- Separation of concerns
- Handlers
- Services
- Repositories
- JSON serialization
- HTTP middleware
- Authentication
- RBAC
- Request validation
- Context propagation
- Testing
- Configuration

**Checkpoint:** Proper CRUD API.

---

### 7. PostgreSQL + OAuth2 Service

Build a service that persists users/OAuth2 information in PostgreSQL.

**Concepts**

- `database/sql`
- PostgreSQL
- Connection pools
- SQL queries
- Transactions
- Migrations
- Repository pattern
- OAuth2
- Sessions/tokens
- Configuration
- Environment variables
- Integration testing

**Checkpoint:** Go application backed by a real database.

---

### 8. Full PWA Backend

Build the backend for a complete PWA.

**Concepts**

- Production API architecture
- Authentication/authorization
- PostgreSQL
- File handling
- Background processing
- SSE/WebSockets
- Caching
- Concurrency
- Structured logging
- Observability
- Graceful shutdown
- Configuration management
- Deployment
- Production error handling

**Checkpoint:** Full production-style Go backend.

---

## Overall Progression

```
Go Fundamentals
       │
       ▼
1. CLI Task Manager
       │
       ▼
2. File Organizer
       │
       ▼
3. Log Analyzer
       │
       ▼
4. HTTP Server
       │
       ▼
5. URL Shortener API
       │
       ▼
6. Task Manager API
       │
       ▼
7. PostgreSQL + OAuth2
       │
       ▼
8. Full PWA Backend
```