Go doesn't have classes. Instead it has structs (data) plus methods (behavior attached to that data) plus interfaces (behavior contracts).

## Structs

A struct is a named collection of fields.

```go
type User struct {
    ID    int
    Name  string
    Email string
}
```

## Struct initialization

There are everal ways to create one, in increasing order of how idiomatic they are:

```go
var u User               // zero-valued: ID 0, Name "", Email ""
u2 := User{1, "Alice", "a@example.com"}          // positional — fragile, avoid
u3 := User{ID: 1, Name: "Alice"}                 // named fields — preferred
u4 := &User{ID: 1, Name: "Alice"}                // pointer to a new struct
```

Named-field initialization is strongly preferred because it doesn't break when reordering or adding fields later.

## Nested structs

Structs can contain other structs, allowing real-world hierarchies to be modelled:

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

Field access uses dot notation, and works the same on a struct or a pointer to one and Go automatically dereferences:

```go
u := User{Name: "Alice"}
u.Name = "Bob"

p := &u
p.Name = "Carol" // no need to write (*p).Name
```

Field names starting with a capital letter are exported (visible outside the package); lowercase ones are private to the package. This is Go's entire visibility system, it has no `public`/`private` keywords.

## Methods

A method is a function with a special "receiver" argument, attaching it to a type:

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

`func (r Rectangle) Area()` takes a **copy** of the struct. Changes made inside the method don't affect the original:

```go
func (r Rectangle) Scale(factor float64) {
    r.Width *= factor // only modifies the copy — no effect outside
}
```

Value receivers are used when the method only reads data, or the struct is small.

## Pointer receivers

`func (r *Rectangle) Scale(...)` takes a pointer, so changes persist:

```go
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

rect := Rectangle{3, 4}
rect.Scale(2) // Go auto-takes the address; works even though rect isn't a pointer
fmt.Println(rect.Width) // 6
```

If any method on a type needs a pointer receiver, *all* methods on that type pointer receivers for consistency. Pointer receivers whenever the method mutates state, or the struct is large enough that copying it is wasteful.

## Interfaces

An interface defines a set of methods a type must have — it's a contract, not a data structure:

```go
type Shape interface {
    Area() float64
}
```

Any type with an `Area() float64` method automatically satisfies `Shape` — nothing else is needed.

## Implicit interface implementation

This is one of the differences from Java/C#: there's no `implements` keyword. If a type has the right methods, it satisfies the interface.

```go
type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

var s Shape = Circle{Radius: 2} // Circle satisfies Shape automatically
```

This allows declaration of small interfaces near where they're *used*, not where the type is defined This can be used to declare interface that a type from a totally different package (including the standard library) satisfies without either side knowing about the other.

## `any` / empty interface

`any` (an alias for `interface{}`) is satisfied by every type — it's Go's escape hatch for "I don't know the type yet":

```go
func describe(v any) {
    fmt.Println(v)
}

describe(42)
describe("hello")
describe(User{Name: "Alice"})
```

Used sparingly as it throws away type safety, so you usually need a type assertion or type switch to do anything useful with the value afterward.

## Type assertions

Extracts the concrete type out of an interface value:

```go
var i any = "hello"

s := i.(string)        // panics if i isn't actually a string
s, ok := i.(string)      // safe form — ok is false instead of panicking
if ok {
    fmt.Println(s)
}
```

The two-value form (`s, ok := ...`) is prepared unless the type is certain and a crash is wanted.

## Type switches

Branch on the concrete type stored inside an interface value:

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

This is the standard way to handle "one of several possible types" without runtime reflection.

## Struct embedding / composition

Go has no inheritance, but one struct (or interface) can be embedded inside another to get field/method promotion. This is Go's substitute for "is-a" relationships:

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

This is composition, not inheritance: `Dog` *has* an `Animal`, and its fields/methods are just promoted to the outer struct for convenience. There's no polymorphism through embedding.