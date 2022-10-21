package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	"github.com/upbound/upbound-go-api3/internal"
	api "github.com/upbound/upbound-go-api3/internal/api/health"
)

// Server is a liveness and readiness server.
func Server(opts internal.CommonOptions) (*http.Server, error) {
	r := chi.NewRouter()
	r.Use(chimid.RedirectSlashes)
	r.Use(chimid.Compress(5))

	api.HandlerFromMux(New(WithLogger(opts.Log)), r)

	return &http.Server{
		Handler:           r,
		Addr:              fmt.Sprintf(":%d", opts.PrivatePort),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}, nil
}
