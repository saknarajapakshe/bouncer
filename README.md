# Bouncer

Bouncer is a Go authorization management library for checking whether a subject
has a permission on a resource.

The core authorization decision is intentionally small:

```go
allowed, err := bouncer.HasPermission(ctx, subjectID, resourceID, permission)
```

Bouncer models authorization with three aspects:

- `subject`: the actor requesting access, usually a user, service account, team,
  or tenant-specific identity.
- `resource`: the object being accessed, such as a document, project, account,
  invoice, or API object.
- `permission`: the action or capability being requested, such as `read`,
  `write`, `delete`, `approve`, or `admin`.

Resources are not treated as a hierarchy by Bouncer. A resource identifier is an
opaque value. If an application wants hierarchical behavior such as
organization -> project -> document inheritance, wildcard resources, or
collision handling between resource names, that logic belongs in the application
layer. Bouncer only answers the permission question for the exact resource ID it
is given.

## Goals

- Provide a focused authorization API for Go services.
- Keep the main check easy to call from application code.
- Support configurable persistence.
- Use native SQL for the initial database implementation.
- Keep the storage layer replaceable so other databases can be added later.
- Optionally expose authorization management endpoints through HTTP.

## Non-Goals

- Bouncer does not define a global resource hierarchy.
- Bouncer does not decide how applications name resources.
- Bouncer does not resolve resource ID collisions.
- Bouncer does not force a specific database vendor.
- Bouncer does not require applications to expose HTTP endpoints.

## Core API

The main interface is expected to look like this:

```go
type Authorizer interface {
    HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error)
}
```

Example usage:

```go
allowed, err := authz.HasPermission(ctx, "user_123", "project_456", "read")
if err != nil {
    return err
}

if !allowed {
    return ErrForbidden
}
```

The library may expose convenience methods around this interface, but
`HasPermission(subjectID, resourceID, permission)` is the primary authorization
decision point.

## Resource Model

Bouncer treats resources as flat, opaque identifiers.

For example, these are all just resource IDs:

```text
project_123
organization_456
organization_456/project_123
invoice:2026:0001
```

Bouncer will not infer that `organization_456/project_123` belongs to
`organization_456`. If an application needs inherited permissions, it can check
multiple resource IDs explicitly:

```go
allowed, err := authz.HasPermission(ctx, subjectID, projectID, "read")
if err != nil {
    return err
}

if !allowed {
    allowed, err = authz.HasPermission(ctx, subjectID, organizationID, "read")
}
```

This keeps the library predictable and leaves application-specific policy
decisions in the application.

## HTTP Endpoints

Bouncer can be used directly as a Go library. If an application needs to expose
authorization management APIs, endpoint registration can be provided through a
function.

The first HTTP integration is expected to use `gorilla/mux`:

```go
router := mux.NewRouter()

err := bouncer.RegisterMuxRoutes(router, service)
if err != nil {
    return err
}
```

Future integrations can add support for other frameworks without changing the
core authorization interface. For example:

- `net/http`
- `chi`
- `gin`
- `echo`
- `fiber`

The HTTP layer should remain an adapter over the same service API used by normal
Go callers. Applications that do not need management endpoints should not have
to register any routes.

## Persistence

The initial implementation should use native SQL through Go's standard
`database/sql` package.

Database configuration should be provided by the application:

```go
db, err := sql.Open("postgres", dsn)
if err != nil {
    return err
}

store := sqlstore.New(db)
authz := bouncer.New(store)
```

This keeps connection ownership with the application. The application decides
which driver to use, how connection pooling is configured, how migrations are
run, and how database lifecycle is managed.

## Database Abstraction Pattern

The most common Go pattern for replaceable persistence is to define a small
interface at the service boundary and provide concrete adapter implementations.
This is usually combined with dependency injection through constructors.

Example:

```go
type Store interface {
    HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error)
    GrantPermission(ctx context.Context, subjectID string, resourceID string, permission string) error
    RevokePermission(ctx context.Context, subjectID string, resourceID string, permission string) error
}

type Service struct {
    store Store
}

func New(store Store) *Service {
    return &Service{store: store}
}
```

The SQL implementation can live behind that interface:

```go
type SQLStore struct {
    db *sql.DB
}

func New(db *sql.DB) *SQLStore {
    return &SQLStore{db: db}
}
```

Later, another database can be plugged in by implementing the same interface:

```go
type RedisStore struct {
    client *redis.Client
}

func (s *RedisStore) HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error) {
    // Redis-specific lookup.
}
```

This approach keeps the domain service independent from the database. It also
makes tests straightforward because test code can provide an in-memory or fake
store.

## Suggested Package Shape

Bouncer should follow a common Go project layout with `cmd`, `pkg`, and
`internal`.

The `internal` directory should be organized by domain. This keeps the
implementation boundaries explicit while preventing other modules from importing
private application code directly.

```text
bouncer/
  cmd/
    bouncer/                    # Optional CLI or server entrypoint.

  pkg/
    bouncer/                    # Public library API.
      authorizer.go             # Core public interfaces.
      service.go                # Public constructor and exported service type.

    httpmux/                    # Public gorilla/mux adapter.
      routes.go                 # Route registration for applications that opt in.

    sqlstore/                   # Public database/sql storage adapter.
      store.go                  # SQL implementation of the storage interface.

  internal/
    authorization/              # Permission checks and authorization rules.
      service.go
      store.go

    subject/                    # Subject-related domain behavior.
      subject.go

    resource/                   # Resource-related domain behavior.
      resource.go

    permission/                 # Permission-related domain behavior.
      permission.go

    http/                       # Internal HTTP handlers used by adapters.
      handlers.go

    storage/
      sql/                      # Native SQL queries and database details.
        queries.go
        migrations/             # Optional SQL migrations.
```

The exact structure can evolve, but the important boundary is that public
packages expose stable integration points while `internal` contains the domain
implementation. HTTP and SQL should remain adapters around the core
authorization service.

## Example

```go
package main

import (
    "context"
    "database/sql"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/your-org/bouncer"
    "github.com/your-org/bouncer/httpmux"
    "github.com/your-org/bouncer/sqlstore"
)

func main() {
    ctx := context.Background()

    db, err := sql.Open("postgres", "postgres://user:pass@localhost/bouncer")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    store := sqlstore.New(db)
    authz := bouncer.New(store)

    allowed, err := authz.HasPermission(ctx, "user_123", "project_456", "read")
    if err != nil {
        panic(err)
    }

    if !allowed {
        panic("forbidden")
    }

    router := mux.NewRouter()
    if err := httpmux.RegisterRoutes(router, authz); err != nil {
        panic(err)
    }

    if err := http.ListenAndServe(":8080", router); err != nil {
        panic(err)
    }
}
```

## Status

Bouncer is in early design and implementation. The intended stable center of the
library is the authorization check:

```go
HasPermission(subjectID, resourceID, permission)
```

Other parts of the library, including route registration helpers and database
adapters, should be built around that interface.
