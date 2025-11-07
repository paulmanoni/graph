package graph

import (
	"encoding/json"
	"fmt"
	"time"
)

// JSONTime is a custom time type that handles flexible JSON time formats.
// It supports both RFC3339 strings and array formats commonly used in some APIs.
//
// Supported input formats:
//   - RFC3339 string: "2024-01-15T14:30:00Z"
//   - Array format: [year, month, day, hour, minute, second, nanosecond]
//   - Minimum array: [2024, 1, 15] (time components default to 0)
//
// Output format:
//   - Always RFC3339 string: "2024-01-15T14:30:00Z"
//
// Example usage:
//
//	type Event struct {
//	    Name      string         `json:"name"`
//	    StartTime graph.JSONTime `json:"startTime"`
//	}
//
//	// Accepts: {"startTime": "2024-01-15T14:30:00Z"}
//	// Accepts: {"startTime": [2024, 1, 15, 14, 30, 0]}
//	// Outputs: {"startTime": "2024-01-15T14:30:00Z"}
type JSONTime time.Time

// MarshalJSON implements json.Marshaler interface.
// Serializes JSONTime to RFC3339 format string.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
// Deserializes from either RFC3339 string or array format.
//
// Array format: [year, month, day, hour?, minute?, second?, nanosecond?]
// Missing time components default to 0. All times are assumed to be UTC.
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	// Check for null
	if string(data) == "null" {
		return nil
	}

	// Check if data is a string
	if data[0] == '"' {
		// Try standard time formats if it's a JSON string
		tt, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
		if err == nil {
			*t = JSONTime(tt)
			return nil
		}
		return err
	}

	// Handle array format [year, month, day, hour, minute, second?, nanosecond?]
	var timeArray []int
	if err := json.Unmarshal(data, &timeArray); err != nil {
		return err
	}

	// Ensure we have at least year, month, day
	if len(timeArray) < 3 {
		return fmt.Errorf("invalid time array format: %v", timeArray)
	}

	// Extract values from array with safe defaults
	year := timeArray[0]
	month := time.Month(timeArray[1])
	day := timeArray[2]

	// Default time values if not provided
	hour := 0
	if len(timeArray) > 3 {
		hour = timeArray[3]
	}

	minute := 0
	if len(timeArray) > 4 {
		minute = timeArray[4]
	}

	second := 0
	if len(timeArray) > 5 {
		second = timeArray[5]
	}

	// Handle nanoseconds if available
	var nsec int
	if len(timeArray) > 6 {
		nsec = timeArray[6]
	}

	// Create the time
	tt := time.Date(year, month, day, hour, minute, second, nsec, time.UTC)
	*t = JSONTime(tt)
	return nil
}

// Time converts JSONTime back to the standard time.Time type.
// This is useful when you need to perform time operations or comparisons.
//
// Example:
//
//	var event Event
//	json.Unmarshal(data, &event)
//	standardTime := event.StartTime.Time()
//	// Now use standard time operations
//	if standardTime.After(time.Now()) { ... }
func (t JSONTime) Time() time.Time {
	return time.Time(t)
}
