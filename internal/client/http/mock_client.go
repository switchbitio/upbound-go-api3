package http

import "net/http"

// MockClient is a mock HTTP client.
type MockClient struct {
	DoFn func(req *http.Request) (*http.Response, error)
}

// Do performs a mock HTTP request by calling the underlying Do function.
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFn(req)
}
