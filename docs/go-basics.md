# 1. Go Basics

## Variables & constants

Go has two ways to declare a variable. `var` is explicit and works everywhere, including at package level:

```go
var name string = "Rudra"
var age = 23          // type inferred
next_age := 24            // short declaration — only works inside functions
```

`:=` is just shorthand for `var x = value` with type inference. You'll use it constantly inside functions. You *can't* use it at package level — there, you need `var`.

Constants are declared with `const` and must be knowable at compile time (no function calls, no runtime values):

```go
const MaxRetries = 3
const Pi = 3.14159
```

Go also has `iota`, a counter used to build enum-like sequences of constants:

```go
type Status int

const (
    StatusPending Status = iota // 0
    StatusActive                // 1
    StatusDone                  // 2
)
```

## Zero values

Unlike some languages, Go never leaves a variable "uninitialized." If a variable is declared without assigning it, it gets a predictable default called the **zero value**:

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

This matters a lot in practice — a struct with unset fields isn't garbage, it's a well-defined value you can safely read.

## Data types

The core built-in types in Go:

```go
bool
string
int, int8, int16, int32, int64
uint, uint8 (byte), uint16, uint32, uint64
float32, float64
rune  // alias for int32, represents a Unicode code point
```

For everyday code, `int`, `float64`, `string`, and `bool` are used unless a specific reason (binary protocols, memory-constrained code) exists to pick a narrower type.

## Type conversion

Go does **not** do implicit type conversion. Mixing an `int` and a `float64` in an expression is a compile error. Explicit conversion:

```go
var i int = 42
var f float64 = float64(i)
var u uint = uint(f)

s := "123"
n, err := strconv.Atoi(s) // string -> int, with error handling
s2 := strconv.Itoa(42)     // int -> string
```

This  prevents a whole class of silent bugs common in languages with coercive typing.

## Operators

Similar to C-family languages:

```go
+ - * / %          // arithmetic
== != < > <= >=      // comparison
&& || !              // logical
& | ^ &^ << >>       // bitwise
= += -= *= /=         // assignment
```

Two Go-specific things to note: there's no ternary operator (`x ? a : b` doesn't exist — use an `if`), and `++`/`--` are statements, not expressions, so `x = y++` doesn’t work.

## Control flow

`if` doesn't need parentheses, and supports an initializer statement scoped to the block:

```go
if err := doSomething(); err != nil {
    return err
}
```

`for` is Go's *only* looping construct — it replaces `while`, `do-while`, and the classic C for-loop:

```go
for i := 0; i < 10; i++ { }   // classic
for condition { }               // while-style
for { }                          // infinite loop
for i, v := range items { }     // range loop
```

`switch` doesn't fall through by default (opposite of C/Java), and doesn't require a condition at all:

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

// Same-type consecutive params can share the type:
func add(a, b int) int {
    return a + b
}
```

## Multiple returns

Go functions can return more than one value — this is the idiomatic way to return "the result, and whether it worked":

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

This pattern is in nearly every standard-library function. Go's substitute for exceptions is `if err != nil` .

## Variadic functions

A function can accept a variable number of arguments using `...`:

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum(1, 2, 3)        // works
sum()                 // also works, nums is an empty slice
nums := []int{1,2,3}
sum(nums...)          // spread a slice into variadic args
```

`fmt.Println(a, b, c)` is itself a variadic function, so it takes any number of arguments.

## Anonymous functions

Functions without a name, often assigned to a variable or passed around immediately:

```go
square := func(x int) int {
    return x * x
}
fmt.Println(square(4)) // 16

// Or invoked immediately:
func() {
    fmt.Println("runs right away")
}()
```

These are used with goroutines and as inline callbacks.

## Function values

Functions in Go are first-class values as they can be stored in variables, passed as arguments, and returned from other functions:

```go
func applyTwice(f func(int) int, x int) int {
    return f(f(x))
}

double := func(x int) int { return x * 2 }
applyTwice(double, 3) // 12
```

This is used in HTTP middleware handlers.

## Pointers

A pointer holds the memory address of a value instead of the value itself. `&` takes the address, `*` dereferences it:

```go
x := 10
p := &x        // p is *int, holding x's address
*p = 20        // change x through the pointer
fmt.Println(x)  // 20
```

The main reason to use a pointer: to let a function modify the caller's original value, or to avoid copying a large struct on every function call.

```go
func increment(n *int) {
    *n++
}
x := 5
increment(&x)
fmt.Println(x) // 6
```

## Arrays

Fixed-size, and the size is part of the type — `[3]int` and `[5]int` are different types entirely. You'll rarely use arrays directly in application code; slices are almost always the better tool.

```go
var a [3]int          // [0 0 0]
b := [3]int{1, 2, 3}
c := [...]int{1, 2, 3} // size inferred from literal
```

## Slices

A slice is a flexible, growable view over an underlying array — this is what you'll actually use for "lists" in Go.

```go
s := []int{1, 2, 3}
s = append(s, 4)         // [1 2 3 4]
sub := s[1:3]              // [2 3] — a view, not a copy!
len(s)                     // length
cap(s)                     // capacity of underlying array

make([]int, 5)             // slice of length 5, zero-valued
make([]int, 0, 10)         // length 0, capacity 10 (pre-allocated)
```

The key mental model: slicing (`s[1:3]`) doesn't copy data — it shares the underlying array. Mutating the sub-slice can mutate the original.

## Maps

Go's built-in hash map:

```go
m := map[string]int{"a": 1, "b": 2}
m["c"] = 3
v, ok := m["a"]       // ok is false if key doesn't exist
delete(m, "a")
for k, v := range m { }  // iteration order is NOT guaranteed

m2 := make(map[string]int) // empty map, ready to use
```

Reading a missing key doesn't panic — it just returns the zero value, so always check `ok` when the presence of a key actually matters.

## Strings

Strings in Go are immutable byte sequences, UTF-8 encoded by default.

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

Because strings are UTF-8, indexing with `s[i]` gives you a raw byte, which can be wrong for non-ASCII text. Use `range` or convert to `[]rune` when you actually need characters.

## Errors

`error` is just a built-in interface with one method:

```go
type error interface {
    Error() string
}
```

Create one with `errors.New` or `fmt.Errorf`:

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

`errors.Is` and `errors.As` let you inspect wrapped errors:

```go
if errors.Is(err, os.ErrNotExist) { }
var pathErr *os.PathError
if errors.As(err, &pathErr) { }
```