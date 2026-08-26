# 3. Packages and Modules

This section is about how Go code is organized into files, folders, and dependency-managed projects.

## Packages

Every Go file belongs to a package, declared at the top of the file. All files in the same directory must belong to the same package:

```go
package main   // executable entry point

package user   // a library package, e.g. in a folder named "user"
```

`package main` is special: it must contain a `func main()`, and it's what makes a program produce an executable rather than a library.

## Imports

Bring in code from other packages:

```go
import "fmt"

import (
    "fmt"
    "net/http"
    "myapp/internal/user" // local module package
)
```

Only imported packages actually used are allowed and an unused import is a compile error, not just a warning. This keeps dependency lists honest.

## Exported / unexported identifiers

There's no `public`/`private` keyword. Capitalization is the visibility rule:

```go
func PublicFunc() {}    // exported — visible outside the package
func privateFunc() {}   // unexported — only visible within the package

type User struct {
    Name string // exported field
    age  int     // unexported field
}
```

This applies to functions, types, struct fields, constants, and variables alike.

## Package aliases

When two imports would collide, or a package name is inconvenient, it is aliased:

```go
import (
    "database/sql"
    postgres "github.com/lib/pq"
)

import (
    "fmt"
    myfmt "myapp/fmt" // avoid clashing with the standard fmt
)
```

## `go.mod`

The file that declares your project as a Go module — its name, its Go version, and its dependencies. Created with:

```bash
go mod init github.com/yourname/yourapp
```

Resulting `go.mod` looks roughly like:

```
module github.com/yourname/yourapp

go 1.22

require (
    github.com/lib/pq v1.10.9
)
```

The module path (`github.com/yourname/yourapp`) is also the import prefix other code (or your own internal packages) will use to reference this project.

## Dependencies

Go dependencies are just other modules, referenced by their source repository path and a version:

```go
import "github.com/lib/pq"
```

You don't need a central package registry like npm — Go fetches directly from source control (GitHub, GitLab, etc.) based on the import path.

## `go get`

Adds or updates a dependency, fetching it and recording it in `go.mod`/`go.sum`:

```bash
go get github.com/lib/pq            # latest version
go get github.com/lib/pq@v1.10.9    # specific version
go get -u ./...                      # update all dependencies
```

## `go mod tidy`

Cleans up `go.mod`/`go.sum` so they exactly match what your code actually imports — adds anything missing, removes anything unused:

```bash
go mod tidy
```

Run after adding/removing imports by hand, or after pulling changes from remote. 

## `go.sum`

A checksum file that locks the exact content-hash of every dependency (and its dependencies), so builds are reproducible and dependencies can't be silently swapped for something malicious. It isn’t edited by hand as`go get`/`go mod tidy` maintain it for automatically. Both `go.mod` and `go.sum` should always be committed to version control.