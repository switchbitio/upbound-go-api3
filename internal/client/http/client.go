package http

import "net/http"

// Client performs HTTP requests.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
