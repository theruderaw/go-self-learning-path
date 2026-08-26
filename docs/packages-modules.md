# 3. Packages & Modules

This section covers how Go code is organized into files, directories, and dependency-managed projects.

## Packages

Every Go file belongs to a package, declared at the top of the file. All files within one directory must share the same package:

```go
package main   // executable entry point

package user   // a library package, e.g. in a folder named "user"
```

`package main` is special: it must contain a `func main()`, and produces an executable rather than a library.

## Imports

Code from other packages is brought in with `import`:

```go
import "fmt"

import (
    "fmt"
    "net/http"
    "myapp/internal/user" // local module package
)
```

An imported package that is not used anywhere in the file is a compile error, not merely a warning — this keeps dependency lists accurate.

## Exported / unexported identifiers

There is no `public`/`private` keyword; capitalization determines visibility:

```go
func PublicFunc() {}    // exported — visible outside the package
func privateFunc() {}   // unexported — visible only within the package

type User struct {
    Name string // exported field
    age  int     // unexported field
}
```

The rule applies uniformly to functions, types, struct fields, constants, and variables.

## Package aliases

When two imports collide, or a package name is inconvenient, an alias is assigned:

```go
import (
    "database/sql"
    postgres "github.com/lib/pq"
)

import (
    "fmt"
    myfmt "myapp/fmt" // avoids clashing with the standard fmt package
)
```

## `go.mod`

Declares a project as a Go module: its name, Go version, and dependencies. Created with:

```bash
go mod init github.com/yourname/yourapp
```

The resulting file resembles:

```
module github.com/yourname/yourapp

go 1.22

require (
    github.com/lib/pq v1.10.9
)
```

The module path also serves as the import prefix used to reference the project's own internal packages.

## Dependencies

Go dependencies are simply other modules, referenced by source-repository path and version:

```go
import "github.com/lib/pq"
```

No central package registry is required — dependencies are fetched directly from source control (GitHub, GitLab, etc.) based on the import path.

## `go get`

Adds or updates a dependency, fetching it and recording it in `go.mod`/`go.sum`:

```bash
go get github.com/lib/pq            # latest version
go get github.com/lib/pq@v1.10.9    # specific version
go get -u ./...                      # update all dependencies
```

## `go mod tidy`

Reconciles `go.mod`/`go.sum` with what the code actually imports — adding anything missing and removing anything unused:

```bash
go mod tidy
```

This is run after imports are added or removed by hand, or after pulling changes made by others.

## `go.sum`

A checksum file locking the exact content hash of every dependency and its own dependencies, ensuring reproducible builds and preventing a dependency from being silently substituted. It is maintained automatically by `go get`/`go mod tidy` and is not edited by hand. Both `go.mod` and `go.sum` are committed to version control.