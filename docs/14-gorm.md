# 14. GORM

GORM is the dominant ORM in the Go ecosystem, offering an alternative to writing raw SQL by hand (as in sections 10 and 13). It trades some directness and performance for significantly less boilerplate around common CRUD operations, associations, and migrations.

## Installing

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

A driver package is required alongside GORM itself, matching whichever database is in use (`postgres`, `mysql`, `sqlite`, `sqlserver`).

## Connecting

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
    log.Fatal("failed to connect database:", err)
}
```

`db` here is a `*gorm.DB`, GORM's equivalent of the connection pool described in sections 10 and 13. It wraps a `database/sql` connection pool internally, which can still be accessed and tuned directly:

```go
sqlDB, err := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(25)
sqlDB.SetConnMaxLifetime(time.Hour)
```

## Models

A model is a plain Go struct, with GORM inferring the table name and column mapping through convention:

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    Email     string    `gorm:"unique"`
    Age       int
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

By convention, `User` maps to a table named `users`, and each exported field maps to a snake_case column (`Email` → `email`). `gorm.Model` is a predefined struct providing `ID`, `CreatedAt`, `UpdatedAt`, and `DeletedAt` (for soft deletes), which can be embedded instead of declaring these fields manually:

```go
type User struct {
    gorm.Model
    Name  string
    Email string
}
```

## Struct tags

GORM tags configure column behavior, similar in spirit to `encoding/json` tags but with different keywords:

```go
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Email    string `gorm:"unique;not null"`
    Age      int    `gorm:"default:18"`
    Bio      string `gorm:"type:text"`
    Internal string `gorm:"-"` // excluded from the database entirely
}
```

## Migrations — `AutoMigrate`

GORM can create or update database tables to match a struct's shape automatically:

```go
db.AutoMigrate(&User{}, &Post{})
```

`AutoMigrate` creates missing tables, adds missing columns, and adds missing indexes — it does not remove columns or change existing column types, to avoid unexpected destructive changes. For production systems with a real migration history, a dedicated migration tool (such as `golang-migrate`, mentioned in section 10's context) is often layered on top rather than relying on `AutoMigrate` alone.

## Create

```go
user := User{Name: "Alice", Email: "alice@example.com"}
result := db.Create(&user)

if result.Error != nil {
    log.Fatal(result.Error)
}
fmt.Println(user.ID) // populated automatically after insertion
```

Multiple records can be created in a single call:

```go
users := []User{{Name: "Bob"}, {Name: "Carol"}}
db.Create(&users)
```

## Find / First / Where

```go
var user User
db.First(&user, 1)                          // primary key = 1
db.First(&user, "email = ?", "a@example.com")

var users []User
db.Find(&users)                              // all rows
db.Where("age > ?", 18).Find(&users)         // conditional
db.Where("name = ? AND age > ?", "Alice", 18).Find(&users)

db.Order("created_at desc").Limit(10).Find(&users)
```

`First` returns an error (`gorm.ErrRecordNotFound`) if no row matches, retrievable the same way as any other error:

```go
err := db.First(&user, 999).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    // no such user
}
```

## Update

```go
db.Model(&user).Update("Name", "Alice Updated")
db.Model(&user).Updates(User{Name: "Alice", Age: 30}) // updates only non-zero fields
db.Model(&user).Updates(map[string]any{"Name": "Alice", "Age": 0}) // map form updates zero values too
```

The struct form of `Updates` skips zero-valued fields (since GORM cannot distinguish "explicitly set to zero" from "left unset"), while the map form updates exactly the fields provided, including zero values.

## Delete

```go
db.Delete(&user, 1)
db.Where("age < ?", 18).Delete(&User{})
```

If the model embeds `gorm.Model` (or otherwise has a `DeletedAt` field), `Delete` performs a **soft delete** by default — it sets `DeletedAt` rather than removing the row, and subsequent queries automatically exclude soft-deleted records:

```go
db.Unscoped().Delete(&user, 1) // permanent deletion, bypassing the soft-delete behavior
db.Unscoped().Find(&users)      // includes soft-deleted records
```

## Associations

GORM recognizes relationships between models based on field types and naming conventions.

**Has One / Belongs To:**

```go
type Profile struct {
    ID     uint
    UserID uint
    Bio    string
}

type User struct {
    ID      uint
    Name    string
    Profile Profile // has one
}
```

**Has Many:**

```go
type Post struct {
    ID     uint
    UserID uint
    Title  string
}

type User struct {
    ID    uint
    Name  string
    Posts []Post // has many
}
```

**Many to Many:**

```go
type Tag struct {
    ID   uint
    Name string
}

type Post struct {
    ID   uint
    Tags []Tag `gorm:"many2many:post_tags;"`
}
```

## Preloading

By default, associated records are not loaded automatically — `Preload` requests them explicitly, avoiding unnecessary queries when they are not needed:

```go
var user User
db.Preload("Posts").First(&user, 1)
// user.Posts is now populated

var users []User
db.Preload("Posts").Preload("Profile").Find(&users)
```

Without `Preload`, `user.Posts` remains an empty slice even if related rows exist in the database.

## Hooks

GORM calls specific methods automatically at points in a record's lifecycle, if they are defined on the model:

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    log.Println("created user:", u.ID)
    return nil
}
```

Common hooks include `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeDelete`, and `AfterDelete`.

## Transactions

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
        return err // triggers a rollback
    }
    if err := tx.Create(&Profile{Bio: "hello"}).Error; err != nil {
        return err
    }
    return nil // commits
})
```

The manual form, comparable to section 10's `db.Begin()`/`tx.Commit()`, is also available:

```go
tx := db.Begin()
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

## Raw SQL

GORM does not require every query to go through its builder — raw SQL can be used directly when convenient:

```go
db.Raw("SELECT id, name FROM users WHERE age > ?", 18).Scan(&users)
db.Exec("UPDATE users SET age = age + 1 WHERE id = ?", 1)
```

## Scopes

A scope is a reusable query fragment, useful for conditions applied repeatedly across an application:

```go
func ActiveUsers(db *gorm.DB) *gorm.DB {
    return db.Where("active = ?", true)
}

func RecentUsers(db *gorm.DB) *gorm.DB {
    return db.Where("created_at > ?", time.Now().AddDate(0, -1, 0))
}

db.Scopes(ActiveUsers, RecentUsers).Find(&users)
```

## Context

Every method has a `WithContext` variant, allowing a `context.Context` (section 8) to control cancellation and timeouts the same way it does with `database/sql` and `pgx`:

```go
db.WithContext(ctx).Find(&users)
db.WithContext(ctx).Create(&user)
```

## GORM vs raw `database/sql` / `pgx`

| | `database/sql` / `pgx` | GORM |
|---|---|---|
| Boilerplate for simple CRUD | more | less |
| Control over exact SQL | full | partial — raw SQL available as an escape hatch |
| Performance | best | slightly slower due to reflection and query building |
| Associations, preloading | manual joins | built in |
| Migrations | external tool | `AutoMigrate`, plus optional external tools |
| Learning curve | lower conceptually, more code | more concepts, less code |

GORM is generally a reasonable choice when development speed on standard CRUD-heavy features matters more than having full control over every generated query. For code with complex, performance-sensitive queries, or a preference for writing SQL directly, `pgx` (section 13) remains the more direct tool — the two are not mutually exclusive within a single project, and it is common to use GORM for routine operations while dropping to raw SQL for specific, complex queries.
