# 9. File I/O

Less central than HTTP/database work for a typical PWA backend, but still needed for config files, uploaded files, logs, or static assets.

## `os`

The package for interacting with the operating system — files, environment variables, process args, exit codes:

```go
import "os"

os.Open("file.txt")        // open for reading
os.Create("file.txt")       // create/truncate for writing
os.ReadFile("file.txt")     // read whole file into memory
os.WriteFile("file.txt", data, 0644) // write whole file at once
os.Remove("file.txt")
os.Getenv("PORT")
os.Exit(1)
```

## `io`

Defines the core streaming interfaces (`Reader`, `Writer`, etc.) that a huge amount of the standard library like files, HTTP bodies, network connections, buffers are built around. 

```go
import "io"

io.Copy(dst, src)      // stream from any Reader to any Writer
io.ReadAll(r)            // read an entire Reader into memory
```

## Reading files

For small/medium files, everything can be read at once:

```go
data, err := os.ReadFile("config.json")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

For large files, files are streamed instead of loading it all into memory:

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

// Or stream writes:
f, err := os.Create("output.txt")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

f.WriteString("hello\n")
f.Write([]byte("more data\n"))
```

`0644` is a Unix permission mode (owner read/write, everyone else read-only) — the standard default for regular files.

## Directories

```go
os.Mkdir("data", 0755)          // single directory
os.MkdirAll("data/uploads", 0755) // nested directories, like `mkdir -p`

entries, err := os.ReadDir(".")
for _, e := range entries {
    fmt.Println(e.Name(), e.IsDir())
}

os.RemoveAll("data") // recursive delete — be careful with this one
```

## `filepath`

Builds and manipulates file paths in an OS-independent way — always prefer this over manually concatenating strings with `/`, since Windows uses `\`:

```go
import "path/filepath"

filepath.Join("data", "uploads", "photo.jpg") // "data/uploads/photo.jpg"
filepath.Ext("photo.jpg")                      // ".jpg"
filepath.Base("data/uploads/photo.jpg")        // "photo.jpg"
filepath.Dir("data/uploads/photo.jpg")         // "data/uploads"
filepath.Abs("data")                            // absolute path
```

## `io.Reader`

An interface — anything with a `Read(p []byte) (n int, err error)` method. This interface is why files, HTTP response bodies, network sockets, and in-memory buffers can all be handled with the same code:

```go
func processData(r io.Reader) error {
    data, err := io.ReadAll(r)
    if err != nil {
        return err
    }
    fmt.Println(len(data))
    return nil
}

// all of these work as an io.Reader:
f, _ := os.Open("file.txt")
processData(f)

resp, _ := http.Get(url)
processData(resp.Body)

processData(strings.NewReader("in-memory string"))
```

## `io.Writer`

The mirror image, anything with `Write(p []byte) (n int, err error)`. `http.ResponseWriter`, `os.File`, and `bytes.Buffer` all satisfy it, which is why `fmt.Fprintf` works identically against a file, an HTTP response, or a buffer:

```go
func logMessage(w io.Writer, msg string) {
    fmt.Fprintf(w, "[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}

logMessage(os.Stdout, "server started")
logMessage(responseWriter, "request handled")

var buf bytes.Buffer
logMessage(&buf, "captured in memory")
```

This pattern of coding against `io.Reader`/`io.Writer` instead of concrete types is a big part of Go's standard library's composition.