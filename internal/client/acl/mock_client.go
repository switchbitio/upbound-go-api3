package acl

import "context"

// MockClient is a mock ACL client.
type MockClient struct {
	GetACLFn func(ctx context.Context, entityID string) (ACL, error)
}

// GetACL calls the underlying GetACLFn.
func (m *MockClient) GetACL(ctx context.Context, entityID string) (ACL, error) {
	return m.GetACLFn(ctx, entityID)
}
