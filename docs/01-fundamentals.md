# 1. Fundamentals

This layer underlies everything else in Go — structs, HTTP handlers, goroutines are all built from these pieces.

## Variables & constants

Go provides two declaration forms. `var` is explicit and works at any scope, including package level:

```go
var name string = "Claude"
var age = 5          // type inferred
age2 := 5             // short declaration — only valid inside functions
```

`:=` is shorthand for `var x = value` with type inference. It is restricted to function bodies; package-level declarations require `var`.

Constants are declared with `const` and must be resolvable at compile time — no function calls, no runtime values:

```go
const MaxRetries = 3
const Pi = 3.14159
```

`iota` generates incrementing values for enum-like constant groups:

```go
type Status int

const (
    StatusPending Status = iota // 0
    StatusActive                // 1
    StatusDone                  // 2
)
```

## Zero values

A declared-but-unassigned variable is never left uninitialized. It receives a predictable default called the **zero value**:

- numbers → `0`
- strings → `""`
- booleans → `false`
- pointers, slices, maps, channels, functions, interfaces → `nil`

```go
var count int      // 0
var label string    // ""
var ok bool         // false
var data []byte      // nil
```

A struct with unset fields is therefore a well-defined, safely readable value rather than garbage memory.

## Data types

The built-in types in common use:

```go
bool
string
int, int8, int16, int32, int64
uint, uint8 (byte), uint16, uint32, uint64
float32, float64
rune  // alias for int32, represents a Unicode code point
```

For most application code, `int`, `float64`, `string`, and `bool` cover the majority of cases; narrower types are reserved for binary protocols or memory-constrained situations.

## Type conversion

Go never performs implicit type conversion. Mixing an `int` and a `float64` in an expression is a compile error. Conversion is always explicit:

```go
var i int = 42
var f float64 = float64(i)
var u uint = uint(f)

s := "123"
n, err := strconv.Atoi(s) // string -> int, with error handling
s2 := strconv.Itoa(42)     // int -> string
```

This design eliminates the class of bugs caused by silent auto-coercion between types.

## Operators

Standard C-family operators:

```go
+ - * / %          // arithmetic
== != < > <= >=      // comparison
&& || !              // logical
& | ^ &^ << >>       // bitwise
= += -= *= /=         // assignment
```

Two Go-specific notes: there is no ternary operator — an `if` statement is used instead — and `++`/`--` are statements, not expressions, so constructs like `x = y++` are invalid.

## Control flow

`if` requires no parentheses and supports an initializer statement scoped to the block:

```go
if err := doSomething(); err != nil {
    return err
}
```

`for` is the only looping construct in Go — it covers the roles filled elsewhere by `while`, `do-while`, and the classic C-style for-loop:

```go
for i := 0; i < 10; i++ { }   // classic
for condition { }               // while-style
for { }                          // infinite loop
for i, v := range items { }     // range loop
```

`switch` does not fall through by default (the opposite of C/Java) and does not require a condition:

```go
switch {
case age < 18:
    fmt.Println("minor")
case age < 65:
    fmt.Println("adult")
default:
    fmt.Println("senior")
}
```

## Functions

Basic shape:

```go
func add(a int, b int) int {
    return a + b
}

// Consecutive parameters of the same type can share one type annotation:
func add(a, b int) int {
    return a + b
}
```

## Multiple returns

A function may return more than one value — the idiomatic way to return both a result and a success/failure indicator:

```go
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 2)
if err != nil {
    log.Fatal(err)
}
```

This pattern appears throughout the standard library, which is why Go code is dense with `if err != nil` checks — it substitutes for exception handling.

## Variadic functions

A function accepts a variable number of arguments using `...`:

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum(1, 2, 3)        // valid
sum()                 // also valid — nums becomes an empty slice
nums := []int{1,2,3}
sum(nums...)          // a slice spread into variadic arguments
```

`fmt.Println(a, b, c)` is itself variadic, which is why it accepts any number of arguments.

## Anonymous functions

Functions without a name, typically assigned to a variable or invoked immediately:

```go
square := func(x int) int {
    return x * x
}
fmt.Println(square(4)) // 16

// Invoked immediately:
func() {
    fmt.Println("runs right away")
}()
```

These appear constantly with goroutines and as inline callbacks.

## Function values

Functions are first-class values in Go — they can be stored in variables, passed as arguments, and returned from other functions:

```go
func applyTwice(f func(int) int, x int) int {
    return f(f(x))
}

double := func(x int) int { return x * 2 }
applyTwice(double, 3) // 12
```

This mechanism underlies the middleware patterns used later in HTTP handling.

## Pointers

A pointer holds a memory address rather than a value. `&` takes an address; `*` dereferences it:

```go
x := 10
p := &x        // p is *int, holding x's address
*p = 20        // modifies x through the pointer
fmt.Println(x)  // 20
```

Pointers allow a function to modify a caller's original value, or avoid copying a large struct on every call:

```go
func increment(n *int) {
    *n++
}
x := 5
increment(&x)
fmt.Println(x) // 6
```

## Arrays

Fixed-size, with the size baked into the type — `[3]int` and `[5]int` are distinct types. Arrays are rarely used directly in application code; slices generally serve better.

```go
var a [3]int          // [0 0 0]
b := [3]int{1, 2, 3}
c := [...]int{1, 2, 3} // size inferred from the literal
```

## Slices

A slice is a flexible, growable view over an underlying array, and is the standard tool for representing lists in Go.

```go
s := []int{1, 2, 3}
s = append(s, 4)         // [1 2 3 4]
sub := s[1:3]              // [2 3] — a view, not a copy
len(s)                     // length
cap(s)                     // capacity of underlying array

make([]int, 5)             // slice of length 5, zero-valued
make([]int, 0, 10)         // length 0, capacity 10 (pre-allocated)
```

Slicing (`s[1:3]`) shares the underlying array rather than copying it, so mutating a sub-slice can mutate the original.

## Maps

Go's built-in hash map type:

```go
m := map[string]int{"a": 1, "b": 2}
m["c"] = 3
v, ok := m["a"]       // ok is false if the key is absent
delete(m, "a")
for k, v := range m { }  // iteration order is not guaranteed

m2 := make(map[string]int) // empty map, ready to use
```

Reading a missing key returns the zero value rather than panicking, so the `ok` form is checked whenever key presence matters.

## Strings

Strings are immutable byte sequences, UTF-8 encoded by default.

```go
s := "hello"
len(s)             // byte length, not necessarily character count
s[0]                // byte value (uint8), not a character
s + " world"        // concatenation
strings.ToUpper(s)
strings.Split(s, ",")
strings.Contains(s, "ell")

for i, r := range s { } // ranges over runes (Unicode code points), not bytes
```

Because strings are UTF-8, indexing with `s[i]` returns a raw byte, which can be incorrect for non-ASCII text. `range` or conversion to `[]rune` is used when characters, rather than bytes, are needed.

## Errors

`error` is a built-in interface with a single method:

```go
type error interface {
    Error() string
}
```

An error value is created with `errors.New` or `fmt.Errorf`:

```go
err := errors.New("something broke")
err2 := fmt.Errorf("failed to process %s: %w", name, err) // %w wraps the original error
```

Idiomatic Go checks errors immediately after the call that might produce them, rather than catching exceptions later:

```go
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("reading config: %w", err)
}
```

`errors.Is` and `errors.As` inspect wrapped errors:

```go
if errors.Is(err, os.ErrNotExist) { }
var pathErr *os.PathError
if errors.As(err, &pathErr) { }
```
