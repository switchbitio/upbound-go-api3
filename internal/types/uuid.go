package types

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

// UUID wraps github.com/google/uuid so we can override the Value() func and store them as
// 16 byte values in SQL databases.
type UUID struct {
	uuid.UUID `json:"uuid,omitempty"`
}

// Value binary value of UUID for database serialization.
func (id UUID) Value() (driver.Value, error) {
	return driver.Value(id.UUID[0:16]), nil
}

// NewUUID factory for creating UUID else Panic
func NewUUID() UUID {
	inner := uuid.New()
	return UUID{UUID: inner}
}

// NewRandomUUID - factory for creating UUID else error
func NewRandomUUID() (UUID, error) {
	inner, err := uuid.NewRandom()
	return UUID{UUID: inner}, err
}

// ParseUUID from string to UUID else error
func ParseUUID(s string) (UUID, error) {
	inner, err := uuid.Parse(s)
	return UUID{UUID: inner}, err
}

// MockAnyUUID used in testing to satify sqlmock.Argument
type MockAnyUUID struct{}

// Match satisfies sqlmock.Argument interface
func (u MockAnyUUID) Match(v driver.Value) bool {
	_, ok := v.([]byte)
	return ok
}
