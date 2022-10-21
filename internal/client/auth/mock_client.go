package auth

import (
	"context"
)

// MockClient is a mock auth manager.
type MockClient struct {
	GetUserIDFn   func(ctx context.Context, token string) (uint, error)
	GetEntityIDFn func(ctx context.Context, token string) (Entity, string, error)
}

// GetUserID calls the underlying GetUserIDFn.
func (m *MockClient) GetUserID(ctx context.Context, token string) (uint, error) {
	return m.GetUserIDFn(ctx, token)
}

// GetEntityID calls the underlying GetEntityIDFn.
func (m *MockClient) GetEntityID(ctx context.Context, token string) (Entity, string, error) {
	return m.GetEntityIDFn(ctx, token)
}
