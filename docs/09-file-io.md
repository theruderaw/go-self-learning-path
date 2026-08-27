# 9. File I/O

Less central than HTTP or database work for a typical PWA backend, but still relevant for config files, uploaded files, logs, or static assets.

## `os`

The package for interacting with the operating system — files, environment variables, process arguments, exit codes:

```go
import "os"

os.Open("file.txt")        // open for reading
os.Create("file.txt")       // create/truncate for writing
os.ReadFile("file.txt")     // read the whole file into memory
os.WriteFile("file.txt", data, 0644) // write the whole file at once
os.Remove("file.txt")
os.Getenv("PORT")
os.Exit(1)
```

## `io`

Defines the core streaming interfaces (`Reader`, `Writer`, etc.) that much of the standard library — files, HTTP bodies, network connections, buffers — is built around. It is rarely imported solely to be used on its own; it is imported because many other APIs are expressed in terms of it.

```go
import "io"

io.Copy(dst, src)      // streams from any Reader to any Writer
io.ReadAll(r)            // reads an entire Reader into memory
```

## Reading files

For small to medium files, the entire contents are read at once:

```go
data, err := os.ReadFile("config.json")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

For large files, the content is streamed rather than loaded entirely into memory:

```go
f, err := os.Open("large.log")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}
```

## Writing files

```go
err := os.WriteFile("output.txt", []byte("hello"), 0644)

// Or streamed writes:
f, err := os.Create("output.txt")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

f.WriteString("hello\n")
f.Write([]byte("more data\n"))
```

`0644` is a Unix permission mode (owner read/write, others read-only) — the standard default for regular files.

## Directories

```go
os.Mkdir("data", 0755)          // single directory
os.MkdirAll("data/uploads", 0755) // nested directories, equivalent to `mkdir -p`

entries, err := os.ReadDir(".")
for _, e := range entries {
    fmt.Println(e.Name(), e.IsDir())
}

os.RemoveAll("data") // recursive delete
```

## `filepath`

Builds and manipulates file paths in an OS-independent way, preferred over manual string concatenation since Windows paths use `\` rather than `/`:

```go
import "path/filepath"

filepath.Join("data", "uploads", "photo.jpg") // "data/uploads/photo.jpg"
filepath.Ext("photo.jpg")                      // ".jpg"
filepath.Base("data/uploads/photo.jpg")        // "photo.jpg"
filepath.Dir("data/uploads/photo.jpg")         // "data/uploads"
filepath.Abs("data")                            // absolute path
```

## `io.Reader`

An interface satisfied by anything with a `Read(p []byte) (n int, err error)` method. This single interface is why files, HTTP response bodies, network sockets, and in-memory buffers can all be handled through identical code:

```go
func processData(r io.Reader) error {
    data, err := io.ReadAll(r)
    if err != nil {
        return err
    }
    fmt.Println(len(data))
    return nil
}

// all of the following satisfy io.Reader:
f, _ := os.Open("file.txt")
processData(f)

resp, _ := http.Get(url)
processData(resp.Body)

processData(strings.NewReader("in-memory string"))
```

## `io.Writer`

The counterpart interface, satisfied by anything with `Write(p []byte) (n int, err error)`. `http.ResponseWriter`, `os.File`, and `bytes.Buffer` all satisfy it, which is why `fmt.Fprintf` behaves identically against a file, an HTTP response, or a buffer:

```go
func logMessage(w io.Writer, msg string) {
    fmt.Fprintf(w, "[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

logMessage(os.Stdout, "server started")
logMessage(responseWriter, "request handled")

var buf bytes.Buffer
logMessage(&buf, "captured in memory")
```

Coding against `io.Reader`/`io.Writer` rather than concrete types is a large part of why the standard library composes so well — functions written once work interchangeably with files, network connections, HTTP bodies, and in-memory buffers.
