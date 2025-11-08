package graph

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Benchmark ExtractBearerToken
func BenchmarkExtractBearerToken(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractBearerToken(req)
	}
}

func BenchmarkExtractBearerToken_NoAuth(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractBearerToken(req)
	}
}

// Benchmark Schema Building
func BenchmarkSchemaBuilder_Simple(b *testing.B) {
	params := SchemaBuilderParams{
		QueryFields: []QueryField{
			getDefaultHelloQuery(),
		},
		MutationFields: []MutationField{
			getDefaultEchoMutation(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewSchemaBuilder(params).Build()
	}
}

func BenchmarkSchemaBuilder_Complex(b *testing.B) {
	type User struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Age       int    `json:"age"`
		IsActive  bool   `json:"isActive"`
		CreatedAt string `json:"createdAt"`
	}

	params := SchemaBuilderParams{
		QueryFields: []QueryField{
			NewResolver[User]("user").
				WithArgs(graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.Int},
				}).
				WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
					return User{ID: 1, Name: "Test User", Email: "test@example.com", Age: 30, IsActive: true}, nil
				}).BuildQuery(),
			NewResolver[[]User]("users").
				AsList().
				WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
					return []User{{ID: 1, Name: "User 1"}}, nil
				}).BuildQuery(),
		},
		MutationFields: []MutationField{
			NewResolver[User]("createUser").
				WithArgs(graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.String},
					"email": &graphql.ArgumentConfig{Type: graphql.String},
				}).
				WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
					return User{ID: 1, Name: "New User"}, nil
				}).BuildMutation(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewSchemaBuilder(params).Build()
	}
}

// Benchmark Query Validation
func BenchmarkValidateGraphQLQuery_SimpleQuery(b *testing.B) {
	query := `{ hello }`
	schema, _ := NewSchemaBuilder(SchemaBuilderParams{
		QueryFields: []QueryField{getDefaultHelloQuery()},
	}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateGraphQLQuery(query, &schema)
	}
}

func BenchmarkValidateGraphQLQuery_ComplexQuery(b *testing.B) {
	query := `{
		user(id: 1) {
			id
			name
			email
			posts {
				id
				title
				comments {
					id
					text
				}
			}
		}
	}`

	type Comment struct {
		ID   int    `json:"id"`
		Text string `json:"text"`
	}

	type Post struct {
		ID       int       `json:"id"`
		Title    string    `json:"title"`
		Comments []Comment `json:"comments"`
	}

	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Posts []Post `json:"posts"`
	}

	schema, _ := NewSchemaBuilder(SchemaBuilderParams{
		QueryFields: []QueryField{
			NewResolver[User]("user").
				WithArgs(graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.Int},
				}).
				WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
					return User{ID: 1, Name: "Test"}, nil
				}).BuildQuery(),
		},
	}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateGraphQLQuery(query, &schema)
	}
}

func BenchmarkValidateGraphQLQuery_DeepQuery(b *testing.B) {
	query := `{
		level1 {
			level2 {
				level3 {
					level4 {
						level5 {
							value
						}
					}
				}
			}
		}
	}`

	schema, _ := NewSchemaBuilder(SchemaBuilderParams{
		QueryFields: []QueryField{getDefaultHelloQuery()},
	}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateGraphQLQuery(query, &schema)
	}
}

func BenchmarkValidateGraphQLQuery_WithAliases(b *testing.B) {
	query := `{
		user1: user(id: 1) { name }
		user2: user(id: 2) { name }
		user3: user(id: 3) { name }
		user4: user(id: 4) { name }
	}`

	schema, _ := NewSchemaBuilder(SchemaBuilderParams{
		QueryFields: []QueryField{getDefaultHelloQuery()},
	}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateGraphQLQuery(query, &schema)
	}
}

// Benchmark Type Registration
func BenchmarkRegisterObjectType(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typeName := "TestType"
		RegisterObjectType(typeName, func() *graphql.Object {
			return graphql.NewObject(graphql.ObjectConfig{
				Name: typeName,
				Fields: graphql.Fields{
					"id":   &graphql.Field{Type: graphql.Int},
					"name": &graphql.Field{Type: graphql.String},
				},
			})
		})
	}
}

// Benchmark Utility Functions
func BenchmarkGetArgString(b *testing.B) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"name": "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetArgString(params, "name")
	}
}

func BenchmarkGetArgInt(b *testing.B) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"age": 30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetArgInt(params, "age")
	}
}

func BenchmarkGetArgBool(b *testing.B) {
	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"active": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetArgBool(params, "active")
	}
}

func BenchmarkGetRootString(b *testing.B) {
	params := graphql.ResolveParams{
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				"token": "abc123",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetRootString(params, "token")
	}
}

func BenchmarkGetRootInfo(b *testing.B) {
	type UserDetails struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	params := graphql.ResolveParams{
		Info: graphql.ResolveInfo{
			RootValue: map[string]interface{}{
				"details": map[string]interface{}{
					"id":   1,
					"name": "Test User",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user UserDetails
		_ = GetRootInfo(params, "details", &user)
	}
}

func BenchmarkGetArg_ComplexType(b *testing.B) {
	type Input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	params := graphql.ResolveParams{
		Args: map[string]interface{}{
			"input": map[string]interface{}{
				"name":  "Test",
				"email": "test@example.com",
				"age":   float64(30),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var input Input
		_ = GetArg(params, "input", &input)
	}
}

// Benchmark HTTP Handler
func BenchmarkNewHTTP_DebugMode(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:      true,
		Playground: true,
	}

	handler := NewHTTP(graphCtx)
	query := `{ hello }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkNewHTTP_WithValidation(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:            false,
		EnableValidation: true,
	}

	handler := NewHTTP(graphCtx)
	query := `{ hello }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkNewHTTP_WithSanitization(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:              false,
		EnableSanitization: true,
	}

	handler := NewHTTP(graphCtx)
	query := `{ invalidField }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkNewHTTP_WithAuth(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG: true,
		UserDetailsFn: func(token string) (interface{}, error) {
			return map[string]interface{}{"id": 1, "name": "User"}, nil
		},
	}

	handler := NewHTTP(graphCtx)
	query := `{ hello }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token-12345")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkNewHTTP_CompleteStack(b *testing.B) {
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	graphCtx := &GraphContext{
		SchemaParams: &SchemaBuilderParams{
			QueryFields: []QueryField{
				NewResolver[User]("user").
					WithArgs(graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{Type: graphql.Int},
					}).
					WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
						return User{ID: 1, Name: "Test User", Email: "test@example.com"}, nil
					}).BuildQuery(),
			},
		},
		DEBUG:              false,
		EnableValidation:   true,
		EnableSanitization: true,
		UserDetailsFn: func(token string) (interface{}, error) {
			return map[string]interface{}{"id": 1, "name": "User"}, nil
		},
	}

	handler := NewHTTP(graphCtx)
	query := `{ user(id: 1) { id name email } }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// Benchmark Resolver Creation
func BenchmarkNewResolver_Simple(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResolver[User]("user").
			WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
				return User{ID: 1, Name: "Test"}, nil
			}).BuildQuery()
	}
}

func BenchmarkNewResolver_WithArgs(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResolver[User]("user").
			WithArgs(graphql.FieldConfigArgument{
				"id":   &graphql.ArgumentConfig{Type: graphql.Int},
				"name": &graphql.ArgumentConfig{Type: graphql.String},
			}).
			WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
				return User{ID: 1, Name: "Test"}, nil
			}).BuildQuery()
	}
}

func BenchmarkNewResolver_AsList(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResolver[[]User]("users").
			AsList().
			WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
				return []User{{ID: 1, Name: "Test"}}, nil
			}).BuildQuery()
	}
}

func BenchmarkNewResolver_AsPaginated(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResolver[User]("users").
			AsPaginated().
			WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
				return PaginatedResponse[User]{
					Items:      []User{{ID: 1, Name: "Test"}},
					TotalCount: 1,
					PageInfo: PageInfo{
						HasNextPage:     false,
						HasPreviousPage: false,
					},
				}, nil
			}).BuildQuery()
	}
}

func BenchmarkNewResolver_WithInputObject(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type CreateUserInput struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResolver[User]("createUser").
			WithInputObject(CreateUserInput{}).
			WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
				return User{ID: 1, Name: "Test"}, nil
			}).BuildMutation()
	}
}

// Benchmark Response Writer Wrapper
func BenchmarkResponseWriterWrapper_Write(b *testing.B) {
	w := httptest.NewRecorder()
	wrapper := newResponseWriterWrapper(w)
	data := []byte(`{"data":{"hello":"world"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = wrapper.Write(data)
		wrapper.body.Reset()
	}
}

func BenchmarkResponseWriterWrapper_SanitizeAndWrite(b *testing.B) {
	data := []byte(`{"errors":[{"message":"Unknown field 'invalidField'. Did you mean 'validField'?"}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrapper := newResponseWriterWrapper(w)
		_, _ = wrapper.Write(data)
		wrapper.sanitizeAndWrite()
	}
}

// Benchmark Middleware
func BenchmarkLoggingMiddleware(b *testing.B) {
	resolver := func(p graphql.ResolveParams) (interface{}, error) {
		return "test", nil
	}

	wrapped := LoggingMiddleware(resolver)
	params := graphql.ResolveParams{
		Info: graphql.ResolveInfo{
			FieldName: "testField",
		},
	}

	b.ResetTimer()
	// Redirect stdout to avoid benchmark noise
	oldStdout := io.Discard
	b.SetBytes(1)
	for i := 0; i < b.N; i++ {
		_, _ = wrapped(params)
		_ = oldStdout
	}
}

func BenchmarkCachedFieldResolver(b *testing.B) {
	resolver := func(p graphql.ResolveParams) (interface{}, error) {
		return "expensive computation", nil
	}

	cached := CachedFieldResolver(
		func(p graphql.ResolveParams) string {
			return "cache-key"
		},
		resolver,
	)

	params := graphql.ResolveParams{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cached(params)
	}
}

// Benchmark Query Depth Calculation
func BenchmarkCalculateQueryDepth_Simple(b *testing.B) {
	query := `{ hello }`

	src := source.NewSource(&source.Source{Body: []byte(query)})
	doc, _ := parser.Parse(parser.ParseParams{Source: src})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateQueryDepth(doc, 0)
	}
}

func BenchmarkCalculateQueryDepth_Complex(b *testing.B) {
	query := `{
		user {
			posts {
				comments {
					author {
						profile
					}
				}
			}
		}
	}`

	src := source.NewSource(&source.Source{Body: []byte(query)})
	doc, _ := parser.Parse(parser.ParseParams{Source: src})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateQueryDepth(doc, 0)
	}
}

// Benchmark Alias Counting
func BenchmarkCountAliases(b *testing.B) {
	query := `{
		user1: user(id: 1) { name }
		user2: user(id: 2) { name }
		user3: user(id: 3) { name }
	}`

	src := source.NewSource(&source.Source{Body: []byte(query)})
	doc, _ := parser.Parse(parser.ParseParams{Source: src})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = countAliases(doc)
	}
}

// Benchmark Complexity Calculation
func BenchmarkCalculateQueryComplexity(b *testing.B) {
	query := `{
		user {
			posts {
				comments {
					text
				}
			}
		}
	}`

	src := source.NewSource(&source.Source{Body: []byte(query)})
	doc, _ := parser.Parse(parser.ParseParams{Source: src})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateQueryComplexity(doc, 1)
	}
}

// Benchmark Context Building
func BenchmarkBuildSchemaFromContext_Default(b *testing.B) {
	graphCtx := &GraphContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buildSchemaFromContext(graphCtx)
	}
}

func BenchmarkBuildSchemaFromContext_WithParams(b *testing.B) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	graphCtx := &GraphContext{
		SchemaParams: &SchemaBuilderParams{
			QueryFields: []QueryField{
				NewResolver[User]("user").
					WithResolver(func(p graphql.ResolveParams) (interface{}, error) {
						return User{ID: 1, Name: "Test"}, nil
					}).BuildQuery(),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buildSchemaFromContext(graphCtx)
	}
}

// Benchmark Full Handler Creation
func BenchmarkNew_Handler(b *testing.B) {
	graphCtx := GraphContext{
		Playground: true,
		DEBUG:      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New(graphCtx)
	}
}

func BenchmarkNew_HandlerWithAuth(b *testing.B) {
	graphCtx := GraphContext{
		Playground: true,
		DEBUG:      true,
		UserDetailsFn: func(token string) (interface{}, error) {
			return map[string]interface{}{"id": 1}, nil
		},
		TokenExtractorFn: ExtractBearerToken,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New(graphCtx)
	}
}

// Parallel benchmarks for concurrent scenarios
func BenchmarkNewHTTP_Parallel(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:      true,
		Playground: true,
	}

	handler := NewHTTP(graphCtx)
	query := `{ hello }`

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			body := bytes.NewBufferString(`{"query":"` + query + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/graphql", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkSchemaBuilder_Parallel(b *testing.B) {
	params := SchemaBuilderParams{
		QueryFields: []QueryField{
			getDefaultHelloQuery(),
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = NewSchemaBuilder(params).Build()
		}
	})
}

func BenchmarkRegisterObjectType_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			typeName := "ConcurrentType"
			RegisterObjectType(typeName, func() *graphql.Object {
				return graphql.NewObject(graphql.ObjectConfig{
					Name: typeName,
					Fields: graphql.Fields{
						"id": &graphql.Field{Type: graphql.Int},
					},
				})
			})
			i++
		}
	})
}

// Benchmark HTTP GET requests
func BenchmarkNewHTTP_GET(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:      true,
		Playground: true,
	}

	handler := NewHTTP(graphCtx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/graphql?query={hello}", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// Benchmark with custom root object function
func BenchmarkNewHTTP_WithCustomRootObject(b *testing.B) {
	graphCtx := &GraphContext{
		DEBUG:      true,
		Playground: true,
		RootObjectFn: func(ctx context.Context, r *http.Request) map[string]interface{} {
			return map[string]interface{}{
				"customData": "test",
			}
		},
	}

	handler := NewHTTP(graphCtx)
	query := `{ hello }`
	body := bytes.NewBufferString(`{"query":"` + query + `"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}