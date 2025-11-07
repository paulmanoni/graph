# graph

A modern, secure GraphQL handler for Go with built-in authentication, validation, and an intuitive builder API.

## Features

- 🚀 **Zero Config Start** - Default hello world schema included
- 🔧 **Fluent Builder API** - Clean, type-safe schema construction
- 🔐 **Built-in Auth** - Automatic Bearer token extraction
- 🛡️ **Security First** - Query depth, complexity, and introspection protection
- 🧹 **Response Sanitization** - Remove field suggestions from errors
- ⚡ **Framework Agnostic** - Works with net/http, Gin, or any framework

Built on top of [graphql-go](https://github.com/graphql-go/graphql).

## Installation

```bash
go get github.com/paulmanoni/graph
```

## Quick Start

### Option 1: Default Schema (Zero Config)

Start immediately with a built-in hello world schema:

```go
package main

import (
    "log"
    "net/http"
    "github.com/paulmanoni/graph"
)

func main() {
    // No schema needed! Includes default hello query & echo mutation
    handler := graph.NewHTTP(&graph.GraphContext{
        Playground: true,
        DEBUG:      true,
    })

    http.Handle("/graphql", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Test it:
```graphql
# Query
{ hello }

# Mutation
mutation { echo(message: "test") }
```

### Option 2: Builder Pattern (Recommended)

Use the fluent builder API for clean schema construction:

```go
package main

import (
    "log"
    "net/http"
    "github.com/paulmanoni/graph"
)

// Define your queries
func getHello() graph.QueryField {
    return graph.NewResolver[string]("hello").
        WithResolver(func(p graph.ResolveParams) (interface{}, error) {
            return "Hello, World!", nil
        }).BuildQuery()
}

func getUser() graph.QueryField {
    return graph.NewResolver[User]("user").
        WithArgs(graph.FieldConfigArgument{
            "id": &graph.ArgumentConfig{
                Type: graph.String,
            },
        }).
        WithResolver(func(p graph.ResolveParams) (interface{}, error) {
            id, _ := graph.GetArgString(p, "id")
            return User{ID: id, Name: "Alice"}, nil
        }).BuildQuery()
}

func main() {
    handler := graph.NewHTTP(&graph.GraphContext{
        SchemaParams: &graph.SchemaBuilderParams{
            QueryFields: []graph.QueryField{
                getHello(),
                getUser(),
            },
            MutationFields: []graph.MutationField{},
        },
        Playground: true,
        DEBUG:      false,
    })

    http.Handle("/graphql", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Option 3: Custom Schema

Bring your own graphql-go schema:

```go
import "github.com/graphql-go/graphql"

schema, _ := graphql.NewSchema(graphql.SchemaConfig{
    Query: graphql.NewObject(graphql.ObjectConfig{
        Name: "Query",
        Fields: graphql.Fields{
            "hello": &graphql.Field{
                Type: graphql.String,
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return "world", nil
                },
            },
        },
    }),
})

handler := graph.NewHTTP(&graph.GraphContext{
    Schema:     &schema,
    Playground: true,
})
```

## Authentication

### Automatic Bearer Token Extraction

Token is automatically extracted from `Authorization: Bearer <token>` header and available in all resolvers:

```go
handler := graph.NewHTTP(&graph.GraphContext{
    SchemaParams: &graph.SchemaBuilderParams{
        QueryFields: []graph.QueryField{
            getProtectedQuery(),
        },
    },

    // Optional: Fetch user details from token
    UserDetailsFn: func(token string) (interface{}, error) {
        // Validate JWT, query database, etc.
        user, err := validateAndGetUser(token)
        return user, err
    },
})
```

Access in resolvers:

```go
func getProtectedQuery() graph.QueryField {
    return graph.NewResolver[User]("me").
        WithResolver(func(p graph.ResolveParams) (interface{}, error) {
            // Get token
            token, err := graph.GetRootString(p, "token")
            if err != nil {
                return nil, fmt.Errorf("authentication required")
            }

            // Get user details (if UserDetailsFn provided)
            var user User
            if err := graph.GetRootInfo(p, "details", &user); err != nil {
                return nil, err
            }

            return user, nil
        }).BuildQuery()
}
```

### Custom Token Extraction

Extract tokens from cookies, custom headers, or query params:

```go
handler := graph.NewHTTP(&graph.GraphContext{
    SchemaParams: &graph.SchemaBuilderParams{...},

    TokenExtractorFn: func(r *http.Request) string {
        // From cookie
        if cookie, err := r.Cookie("auth_token"); err == nil {
            return cookie.Value
        }

        // From custom header
        if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
            return apiKey
        }

        // From query param
        return r.URL.Query().Get("token")
    },

    UserDetailsFn: func(token string) (interface{}, error) {
        return getUserByToken(token)
    },
})
```

## Security Features

### Production Setup

Enable all security features for production:

```go
handler := graph.NewHTTP(&graph.GraphContext{
    SchemaParams:       &graph.SchemaBuilderParams{...},
    DEBUG:              false,  // Enable security features
    EnableValidation:   true,   // Validate queries
    EnableSanitization: true,   // Sanitize errors
    Playground:         false,  // Disable playground

    UserDetailsFn: func(token string) (interface{}, error) {
        return validateAndGetUser(token)
    },
})
```

### Validation Rules (when `EnableValidation: true`)

- **Max Query Depth**: 10 levels
- **Max Aliases**: 4 per query
- **Max Complexity**: 200
- **Introspection**: Disabled (blocks `__schema` and `__type`)

### Response Sanitization (when `EnableSanitization: true`)

Removes field suggestions from error messages:

**Before:**
```json
{
  "errors": [{
    "message": "Cannot query field \"nam\". Did you mean \"name\"?"
  }]
}
```

**After:**
```json
{
  "errors": [{
    "message": "Cannot query field \"nam\"."
  }]
}
```

### Debug Mode

Use `DEBUG: true` during development to skip all validation and sanitization:

```go
handler := graph.NewHTTP(&graph.GraphContext{
    SchemaParams: &graph.SchemaBuilderParams{...},
    DEBUG:        true,  // Disables validation & sanitization
    Playground:   true,  // Enable playground for testing
})
```

## Helper Functions

### Extracting Arguments

```go
// String argument
name, err := graph.GetArgString(p, "name")

// Int argument
age, err := graph.GetArgInt(p, "age")

// Bool argument
active, err := graph.GetArgBool(p, "active")

// Complex type
var input CreateUserInput
err := graph.GetArg(p, "input", &input)
```

### Accessing Root Values

```go
// Get token
token, err := graph.GetRootString(p, "token")

// Get user details
var user User
err := graph.GetRootInfo(p, "details", &user)
```

## Framework Integration

### With Gin

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/paulmanoni/graph"
)

func main() {
    r := gin.Default()

    handler := graph.NewHTTP(&graph.GraphContext{
        SchemaParams:     &graph.SchemaBuilderParams{...},
        EnableValidation: true,
    })

    r.POST("/graphql", gin.WrapF(handler))
    r.GET("/graphql", gin.WrapF(handler))

    r.Run(":8080")
}
```

### With Chi

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/paulmanoni/graph"
)

func main() {
    r := chi.NewRouter()

    handler := graph.NewHTTP(&graph.GraphContext{
        SchemaParams: &graph.SchemaBuilderParams{...},
    })

    r.Handle("/graphql", handler)

    http.ListenAndServe(":8080", r)
}
```

### With Standard net/http

```go
handler := graph.NewHTTP(&graph.GraphContext{
    SchemaParams: &graph.SchemaBuilderParams{...},
})

http.Handle("/graphql", handler)
http.ListenAndServe(":8080", nil)
```

## API Reference

### `NewHTTP(graphCtx *GraphContext) http.HandlerFunc`

Creates a standard HTTP handler with validation and sanitization support.

### `GraphContext` Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Schema` | `*graphql.Schema` | `nil` | Custom GraphQL schema (Option 3) |
| `SchemaParams` | `*SchemaBuilderParams` | `nil` | Builder params (Option 2) |
| `Playground` | `bool` | `false` | Enable GraphQL Playground |
| `Pretty` | `bool` | `false` | Pretty-print JSON responses |
| `DEBUG` | `bool` | `false` | Skip validation/sanitization |
| `EnableValidation` | `bool` | `false` | Enable query validation |
| `EnableSanitization` | `bool` | `false` | Enable error sanitization |
| `TokenExtractorFn` | `func(*http.Request) string` | Bearer token | Custom token extraction |
| `UserDetailsFn` | `func(string) (interface{}, error)` | `nil` | Fetch user from token |
| `RootObjectFn` | `func(context.Context, *http.Request) map[string]interface{}` | `nil` | Custom root setup |

**Note:** If both `Schema` and `SchemaParams` are `nil`, a default hello world schema is used.

### `SchemaBuilderParams`

```go
type SchemaBuilderParams struct {
    QueryFields    []QueryField
    MutationFields []MutationField
}
```

## Examples

See the [examples](./examples) directory for complete working examples:

- `main.go` - Full example with authentication

## License

MIT