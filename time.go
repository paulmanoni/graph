package graph

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// SpringShortLayout is the time format used by Spring Boot for DateTime serialization.
// Format: yyyy-MM-dd'T'HH:mm (e.g., "2024-01-15T14:30")
const SpringShortLayout = "2006-01-02T15:04"

// serializeDateTime converts time.Time to Spring Boot compatible string format.
// Always returns time in UTC.
func serializeDateTime(value interface{}) interface{} {
	if t, ok := value.(time.Time); ok {
		// always UTC to match Spring Boot style
		return t.UTC().Format(SpringShortLayout)
	}
	if t, ok := value.(*time.Time); ok && t != nil {
		return t.UTC().Format(SpringShortLayout)
	}
	return nil
}

// unserializeDateTime parses a Spring Boot formatted date string into time.Time.
// Returns time in UTC or nil if parsing fails.
func unserializeDateTime(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		if t, err := time.Parse(SpringShortLayout, s); err == nil {
			return t.UTC()
		}
	}
	return nil
}

// DateTime is a GraphQL scalar type for date-time values.
// It uses the Spring Boot date format: yyyy-MM-dd'T'HH:mm (e.g., "2024-01-15T14:30").
// All times are automatically converted to UTC.
//
// Usage in struct fields:
//
//	type Event struct {
//	    Name      string    `json:"name"`
//	    StartTime time.Time `json:"startTime"` // Will use DateTime scalar
//	}
//
// The scalar automatically handles:
//   - Serialization: time.Time → "2024-01-15T14:30"
//   - Deserialization: "2024-01-15T14:30" → time.Time
//   - UTC conversion for all values
var DateTime = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "DateTime",
	Description: "The `DateTime` scalar type formatted as yyyy-MM-dd'T'HH:mm",
	Serialize:   serializeDateTime,
	ParseValue:  unserializeDateTime,
	ParseLiteral: func(valueAST ast.Value) interface{} {
		if v, ok := valueAST.(*ast.StringValue); ok {
			return unserializeDateTime(v.Value)
		}
		return nil
	},
})
