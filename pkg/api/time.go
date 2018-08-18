package api

import (
	"fmt"
	"time"
)

// JSONTime handles parsing and formatting timestamps according the ISO8601 standard
type JSONTime struct {
	time.Time
}

// MarshalJSON formats the timestamp as JSON
func (t JSONTime) MarshalJSON() ([]byte, error) {
	date := fmt.Sprintf("%q", t.Format(time.RFC3339))
	return []byte(date), nil
}
