# 2. Structs, Methods & Interfaces

Go has no classes. Instead there are structs (data), methods (behavior attached to that data), and interfaces (behavior contracts). Together these replace most of what object-oriented languages achieve with classes.

## Structs

A struct is a named collection of fields, used to group related data under one type:

```go
type User struct {
    ID    int
    Name  string
    Email string
}
```

## Struct initialization

Several forms exist, in increasing order of idiomatic preference:

```go
var u User               // zero-valued: ID 0, Name "", Email ""
u2 := User{1, "Alice", "a@example.com"}          // positional — fragile
u3 := User{ID: 1, Name: "Alice"}                 // named fields — preferred
u4 := &User{ID: 1, Name: "Alice"}                // pointer to a new struct
```

Named-field initialization is preferred because it does not break when fields are reordered or added later.

## Nested structs

Structs may contain other structs, modeling hierarchical data:

```go
type Address struct {
    City    string
    Country string
}

type User struct {
    Name    string
    Address Address
}

u := User{Name: "Alice", Address: Address{City: "Ranchi", Country: "India"}}
fmt.Println(u.Address.City)
```

## Struct fields

Field access uses dot notation, identically whether the value is a struct or a pointer to one — Go dereferences automatically:

```go
u := User{Name: "Alice"}
u.Name = "Bob"

p := &u
p.Name = "Carol" // equivalent to (*p).Name
```

A field name beginning with a capital letter is exported (visible outside its package); a lowercase name is private to the package. This capitalization rule is Go's entire visibility system — there are no `public`/`private` keywords.

## Methods

A method is a function with an additional "receiver" argument, attaching it to a type:

```go
type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

rect := Rectangle{Width: 3, Height: 4}
rect.Area() // 12
```

## Value receivers

`func (r Rectangle) Area()` operates on a **copy** of the struct. Modifications inside the method do not affect the original:

```go
func (r Rectangle) Scale(factor float64) {
    r.Width *= factor // modifies only the copy — no external effect
}
```

Value receivers are appropriate when a method only reads data, or the struct is small.

## Pointer receivers

`func (r *Rectangle) Scale(...)` operates through a pointer, so changes persist:

```go
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

rect := Rectangle{3, 4}
rect.Scale(2) // Go takes the address automatically, even though rect isn't a pointer
fmt.Println(rect.Width) // 6
```

A common convention: if any method on a type requires a pointer receiver, all methods on that type use pointer receivers, for consistency. Pointer receivers are used whenever a method mutates state, or the struct is large enough that copying is wasteful.

## Interfaces

An interface defines a set of methods a type must implement — a contract, not a data structure:

```go
type Shape interface {
    Area() float64
}
```

Any type with an `Area() float64` method satisfies `Shape` automatically.

## Implicit interface implementation

Go has no `implements` keyword. If a type has the required methods, it satisfies the interface — nothing further is declared:

```go
type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

var s Shape = Circle{Radius: 2} // Circle satisfies Shape automatically
```

This allows an interface to be defined near its point of use rather than alongside the implementing type, and permits a type from an entirely unrelated package — including the standard library — to satisfy an interface without either side referencing the other.

## `any` / empty interface

`any` (an alias for `interface{}`) is satisfied by every type — Go's mechanism for "type unknown at this point":

```go
func describe(v any) {
    fmt.Println(v)
}

describe(42)
describe("hello")
describe(User{Name: "Alice"})
```

Because it discards type safety, a value stored as `any` typically requires a type assertion or type switch before further use.

## Type assertions

Extracts the concrete type from an interface value:

```go
var i any = "hello"

s := i.(string)        // panics if i is not a string
s, ok := i.(string)      // safe form — ok is false instead of panicking
if ok {
    fmt.Println(s)
}
```

The two-value form is preferred except when a mismatch should be treated as a fatal error.

## Type switches

Branches on the concrete type stored inside an interface value:

```go
func describe(v any) {
    switch x := v.(type) {
    case int:
        fmt.Println("int:", x)
    case string:
        fmt.Println("string:", x)
    case User:
        fmt.Println("user:", x.Name)
    default:
        fmt.Println("unknown type")
    }
}
```

This is the standard mechanism for handling several possible types without runtime reflection.

## Struct embedding / composition

Go has no inheritance, but a struct or interface may be embedded inside another to promote its fields and methods — Go's substitute for "is-a" relationships:

```go
type Animal struct {
    Name string
}

func (a Animal) Describe() string {
    return "I am " + a.Name
}

type Dog struct {
    Animal      // embedded, no field name
    Breed string
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Labrador"}
d.Describe() // "I am Rex" — promoted from Animal
d.Name       // "Rex" — also promoted
```

This is composition rather than inheritance: `Dog` contains an `Animal`, and its fields and methods are promoted to the outer struct for convenience. No polymorphism results from embedding — a `Dog` cannot be passed where an `Animal` is expected without an explicit conversion.
