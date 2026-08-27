# Go Self-Learning Documentation

The goal of this learning document is to give developers a practical foundation in Go for building a real PWA backend. Each stage introduces concepts that are immediately reinforced through a project.

## Learning Path

1. [Go Basics](docs/01-fundamentals.md)
2. [Structs, Methods & Interfaces](docs/02-structs-methods-interfaces.md)
3. [Packages & Modules](docs/03-packages-modules.md)
4. [Defer & Panic](docs/04-defer-panic.md)
5. [Concurrency](docs/05-concurrency.md)
6. [HTTP / Web](docs/06-http-web.md)
7. [JSON & Data](docs/07-json-data.md)
8. [Context](docs/08-context.md)
9. [File I/O](docs/09-file-io.md)
10. [Database (`database/sql`)](docs/10-database.md)
11. [Testing](docs/11-testing.md)
12. [Project Structure & Tooling](docs/12-project-structure-tooling.md)
13. [pgx (PostgreSQL Driver, In Depth)](docs/13-pgx.md)
14. [GORM](docs/14-gorm.md)
15. [Authentication & Authorization (OAuth2)](docs/15-oauth2.md)

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
- pgx (native API)
- GORM
- PostgreSQL
- Connection pools
- SQL queries
- Transactions
- Migrations
- Repository pattern
- OAuth2 / authentication & authorization
- JWTs, sessions, refresh tokens
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
